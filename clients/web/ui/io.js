import { ansiToHtml } from './ansi.js';

// Approximate the Go server's small palette (Outln color arg 0..7).
// Tuned for dark-background readability; fine-tune later if needed.
const PALETTE = {
  0: '#000000',
  1: '#cd0000',
  2: '#00cd00',
  3: '#cdcd00',
  4: '#5c5cff',
  5: '#cd00cd',
  6: '#00cdcd',
  7: '#e5e5e5',
};

export class Io {
  constructor(rootEl) {
    this.root = rootEl;
  }

  clear() {
    this.root.innerHTML = '';
  }

  cr() {
    this.root.appendChild(document.createElement('br'));
  }

  println(text, color = 7) {
    const span = document.createElement('div');
    span.textContent = text;
    span.style.color = PALETTE[color] ?? PALETTE[7];
    this.root.appendChild(span);
  }

  async showAnsiFile(name) {
    const res = await fetch(`/data/ansi/${name}.ans`);
    if (!res.ok) throw new Error(`ansi fetch ${name}: ${res.status}`);
    // The .ans files store ESC as the literal two-character sequence \e —
    // restore it to 0x1B before handing to the ANSI parser.
    const text = (await res.text()).replace(/\\e/g, '\x1b');
    const pre = document.createElement('pre');
    pre.innerHTML = ansiToHtml(text);
    this.root.appendChild(pre);
  }

  // Wait for the user to press one of the allowed letters (case-insensitive).
  // Mirrors slyio.LettersPrompt: prints the prompt, then resolves with the
  // uppercase character that was pressed.
  lettersPrompt(prompt, allowed) {
    const span = document.createElement('div');
    span.className = 'prompt cursor';
    span.textContent = prompt + ' ';
    this.root.appendChild(span);
    span.scrollIntoView({ block: 'end' });

    const allowedSet = new Set(allowed.toUpperCase().split(''));
    return new Promise((resolve) => {
      const onKey = (e) => {
        const ch = (e.key || '').toUpperCase();
        if (!allowedSet.has(ch)) return;
        e.preventDefault();
        document.removeEventListener('keydown', onKey);
        span.classList.remove('cursor');
        span.textContent = prompt + ' ' + ch;
        resolve(ch);
      };
      document.addEventListener('keydown', onKey);
    });
  }
}
