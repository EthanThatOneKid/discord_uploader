import { config } from "./deps.ts";

config({ export: true });

export const DISCORD_WEBHOOK_URL = Deno.env.get("DISCORD_WEBHOOK_URL") || "";
export const PORT = parseInt(Deno.env.get("PORT") || "8080");
