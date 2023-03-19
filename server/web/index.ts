export const INDEX = Deno.readTextFileSync(
  new URL("index.html", import.meta.url),
);
