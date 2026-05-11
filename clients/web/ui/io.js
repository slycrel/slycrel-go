import { ansiToHtml } from './ansi.js';

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
    const div = document.createElement('div');
    div.textContent = text;
    div.style.color = PALETTE[color] ?? PALETTE[7];
    this.root.appendChild(div);
  }

  async showAnsiFile(name) {
    const res = await fetch(`/data/ansi/${name}.ans`);
    if (!res.ok) throw new Error(`ansi fetch ${name}: ${res.status}`);
    const text = (await res.text()).replace(/\\e/g, '\x1b');
    const pre = document.createElement('pre');
    pre.innerHTML = ansiToHtml(text);
    this.root.appendChild(pre);
  }

  // Single-key prompt with a closed allowed set. Resolves uppercase.
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

  // Free-text input. Resolves on Enter; supports backspace; clamps to maxLen.
  // Empty input returns ''.
  textPrompt(promptText, maxLen = 40) {
    const line = document.createElement('div');
    line.className = 'prompt';
    this.root.appendChild(line);

    const promptSpan = document.createElement('span');
    promptSpan.textContent = promptText + ' ';
    line.appendChild(promptSpan);

    const inputSpan = document.createElement('span');
    line.appendChild(inputSpan);

    const cursor = document.createElement('span');
    cursor.className = 'cursor';
    line.appendChild(cursor);
    line.scrollIntoView({ block: 'end' });

    return new Promise((resolve) => {
      let buffer = '';
      const onKey = (e) => {
        if (e.key === 'Enter') {
          e.preventDefault();
          document.removeEventListener('keydown', onKey);
          cursor.remove();
          resolve(buffer);
        } else if (e.key === 'Backspace') {
          e.preventDefault();
          buffer = buffer.slice(0, -1);
          inputSpan.textContent = buffer;
        } else if (e.key.length === 1 && buffer.length < maxLen) {
          e.preventDefault();
          buffer += e.key;
          inputSpan.textContent = buffer;
        }
      };
      document.addEventListener('keydown', onKey);
    });
  }

  // Mirrors slyio.PausePrompt — any printable key (or Enter/Space) continues.
  pausePrompt(promptText = '--press a key--') {
    const span = document.createElement('div');
    span.className = 'prompt cursor';
    span.textContent = promptText + ' ';
    this.root.appendChild(span);
    span.scrollIntoView({ block: 'end' });

    return new Promise((resolve) => {
      const onKey = (e) => {
        if (e.key.length !== 1 && e.key !== 'Enter') return;
        e.preventDefault();
        document.removeEventListener('keydown', onKey);
        span.classList.remove('cursor');
        resolve();
      };
      document.addEventListener('keydown', onKey);
    });
  }
}
