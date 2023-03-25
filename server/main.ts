import { serve } from "./deps.ts";
import { INDEX } from "./web/mod.ts";
import { DISCORD_WEBHOOK_URL, PORT } from "./env.ts";

if (import.meta.main) {
  await serve(handle, {
    port: PORT,
    onListen({ port }) {
      console.log(`Server running on http://localhost:${port}`);
    },
  });
}

async function handle(request: Request): Promise<Response> {
  const { pathname } = new URL(request.url);
  switch (pathname) {
    case "/": {
      if (request.method === "GET") {
        return new Response(INDEX, {
          headers: { "content-type": "text/html" },
        });
      }

      const webhookArgs = await makeWebhookArgs(DISCORD_WEBHOOK_URL, request);
      const response = await fetch(...webhookArgs);
      const body = await response.text();
      return new Response(body, {
        headers: { "content-type": "application/json" },
      });
    }

    default: {
      return new Response("Not Found", { status: 404 });
    }
  }
}

/**
 * Converts a request to the arguments for a Discord webhook invocation.
 *
 * @see https://discord.com/developers/docs/reference#uploading-files
 */
async function makeWebhookArgs(
  webhookURL: string,
  request: Request,
): Promise<[URL, RequestInit]> {
  const data = await request.formData();
  const file = data.get("file") as File;
  const payload = new FormData();
  payload.append(
    "payload_json",
    new Blob(
      [
        JSON.stringify({
          "embeds": [{
            "image": {
              "url": `attachment://${file.name}`,
            },
          }],
          attachments: [{ id: 0, filename: file.name }],
        }),
      ],
      { type: "application/json" },
    ),
  );
  payload.append("files[0]", file, file.name);

  const url = new URL(webhookURL);
  url.searchParams.set("wait", "true");

  const requestInit: RequestInit = {
    method: "POST",
    body: payload,
    headers: new Headers({
      "Content-Type": "multipart/form-data",
    }),
  };

  return [url, requestInit];
}
