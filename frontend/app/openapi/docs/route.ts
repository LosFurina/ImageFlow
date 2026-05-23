import { NextRequest, NextResponse } from "next/server";
import { buildRuntimeConfig } from "../../utils/serverRuntimeConfig";

export function GET(request: NextRequest) {
  const config = buildRuntimeConfig(request);
  return NextResponse.redirect(config.openapiDocsUrl, 307);
}
