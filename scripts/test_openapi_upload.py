#!/usr/bin/env python3
"""Test ImageFlow OpenAPI image upload with AK/SK HMAC signing.

Usage:
  export IMAGEFLOW_AK='your-access-key'
  export IMAGEFLOW_SK='your-secret-key'
  python3 scripts/test_openapi_upload.py ./example.jpg \
    --base-url http://127.0.0.1:8686 \
    --tags test,openapi \
    --expiry-minutes 60

The script signs the exact multipart request body sent to /openapi/upload.
"""

from __future__ import annotations

import argparse
import hashlib
import hmac
import json
import mimetypes
import os
import sys
import time
import uuid
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import urljoin, urlparse
from urllib.request import Request, urlopen


def build_multipart_body(image_paths: list[Path], tags: str | None, expiry_minutes: int | None) -> tuple[bytes, str]:
    """Build a multipart/form-data body matching backend UploadHandler fields."""
    boundary = f"----ImageFlowOpenAPITest{uuid.uuid4().hex}"
    chunks: list[bytes] = []

    def add_field(name: str, value: str) -> None:
        chunks.append(f"--{boundary}\r\n".encode())
        chunks.append(f'Content-Disposition: form-data; name="{name}"\r\n\r\n'.encode())
        chunks.append(value.encode())
        chunks.append(b"\r\n")

    def add_file(field_name: str, path: Path) -> None:
        content_type = mimetypes.guess_type(path.name)[0] or "application/octet-stream"
        chunks.append(f"--{boundary}\r\n".encode())
        chunks.append(
            (
                f'Content-Disposition: form-data; name="{field_name}"; '
                f'filename="{path.name}"\r\n'
                f"Content-Type: {content_type}\r\n\r\n"
            ).encode()
        )
        chunks.append(path.read_bytes())
        chunks.append(b"\r\n")

    # Backend currently reads files from MultipartForm.File["images[]"].
    for image_path in image_paths:
        add_file("images[]", image_path)

    if tags:
        add_field("tags", tags)

    # Backend currently reads expiry from form field "expiryMinutes".
    if expiry_minutes is not None:
        add_field("expiryMinutes", str(expiry_minutes))

    chunks.append(f"--{boundary}--\r\n".encode())
    return b"".join(chunks), f"multipart/form-data; boundary={boundary}"


def build_signature(secret_key: str, method: str, path: str, timestamp: str, body: bytes) -> tuple[str, str]:
    body_hash = hashlib.sha256(body).hexdigest()
    string_to_sign = f"{method}\n{path}\n{timestamp}\n{body_hash}"
    signature = hmac.new(secret_key.encode(), string_to_sign.encode(), hashlib.sha256).hexdigest()
    return signature, string_to_sign


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Upload image(s) through ImageFlow /openapi/upload using AK/SK.")
    parser.add_argument("images", nargs="+", help="Image file path(s) to upload")
    parser.add_argument("--base-url", default=os.getenv("IMAGEFLOW_BASE_URL", "http://127.0.0.1:8686"), help="Backend base URL, default: %(default)s")
    parser.add_argument("--ak", default=os.getenv("IMAGEFLOW_AK"), help="Access Key, or env IMAGEFLOW_AK")
    parser.add_argument("--sk", default=os.getenv("IMAGEFLOW_SK"), help="Secret Key, or env IMAGEFLOW_SK")
    parser.add_argument("--tags", default="openapi-test", help="Comma-separated tags, default: %(default)s")
    parser.add_argument("--expiry-minutes", type=int, default=0, help="Expiration minutes; 0 means never expire, default: %(default)s")
    parser.add_argument("--timeout", type=int, default=60, help="HTTP timeout seconds, default: %(default)s")
    parser.add_argument("--verbose", action="store_true", help="Print signed request metadata, without printing SK")
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    if not args.ak or not args.sk:
        print("ERROR: missing AK/SK. Use --ak/--sk or env IMAGEFLOW_AK/IMAGEFLOW_SK.", file=sys.stderr)
        return 2

    image_paths = [Path(p).expanduser().resolve() for p in args.images]
    missing = [str(p) for p in image_paths if not p.is_file()]
    if missing:
        print("ERROR: image file not found:", file=sys.stderr)
        for p in missing:
            print(f"  - {p}", file=sys.stderr)
        return 2

    base_url = args.base_url.rstrip("/") + "/"
    url = urljoin(base_url, "openapi/upload")
    parsed = urlparse(url)
    path_for_signing = parsed.path  # Backend signs r.URL.Path only, no query string.

    body, content_type = build_multipart_body(image_paths, args.tags, args.expiry_minutes)
    timestamp = str(int(time.time()))
    signature, string_to_sign = build_signature(args.sk, "POST", path_for_signing, timestamp, body)

    headers = {
        "Content-Type": content_type,
        "Content-Length": str(len(body)),
        "X-Access-Key": args.ak,
        "X-Timestamp": timestamp,
        "X-Signature": signature,
    }

    if args.verbose:
        print(f"POST {url}")
        print(f"AK: {args.ak}")
        print(f"Timestamp: {timestamp}")
        print(f"Body bytes: {len(body)}")
        print("StringToSign:")
        print(string_to_sign)
        print(f"Signature: {signature}")
        print()

    request = Request(url, data=body, headers=headers, method="POST")

    try:
        with urlopen(request, timeout=args.timeout) as response:
            status = response.status
            response_body = response.read().decode("utf-8", errors="replace")
    except HTTPError as exc:
        status = exc.code
        response_body = exc.read().decode("utf-8", errors="replace")
    except URLError as exc:
        print(f"ERROR: request failed: {exc}", file=sys.stderr)
        return 1

    print(f"HTTP {status}")
    try:
        print(json.dumps(json.loads(response_body), ensure_ascii=False, indent=2))
    except json.JSONDecodeError:
        print(response_body)

    return 0 if 200 <= status < 300 else 1


if __name__ == "__main__":
    raise SystemExit(main())
