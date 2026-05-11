// ansi_up converts ANSI escape sequences (color, etc.) into styled HTML.
// Using the jsdelivr CDN ESM build keeps this client zero-build; if we
// ever want offline-safe deploys, vendor ansi_up.js under clients/web/vendor/
// and swap the import.
import AnsiUp from 'https://cdn.jsdelivr.net/npm/ansi_up@6/+esm';

const converter = new AnsiUp();
converter.use_classes = false;
converter.escape_html = true;

export function ansiToHtml(text) {
  return converter.ansi_to_html(text);
}
