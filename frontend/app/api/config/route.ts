import { NextRequest, NextResponse } from "next/server";
import { buildRuntimeConfig } from "../../utils/serverRuntimeConfig";

export async function GET(request: NextRequest) {
  return NextResponse.json(buildRuntimeConfig(request));
}
