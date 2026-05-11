import { ansiToHtml } from './ansi.js';

// Mirrors internal/io/ansi.go's colorCodes — the original Slycrel palette
// uses 1=cyan (info/prompts), 2=bright white, 3=bright green, 4=bright
// yellow (character names), 5=bright magenta, 6=bright red (errors).
// Routing through ansi_up keeps println output visually consistent with
// the BBS-art renderer (which also speaks ANSI escape codes).
const COLOR_TO_ANSI = {
  0: '\x1b[0m',
  1: '\x1b[0;36m',
  2: '\x1b[1;37m',
  3: '\x1b[1;32m',
  4: '\x1b[1;33m',
  5: '\x1b[1;35m',
  6: '\x1b[1;31m',
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

  println(text, color) {
    const wrapped = color !== undefined
      ? (COLOR_TO_ANSI[color] ?? '') + text + '\x1b[0m'
      : text;
    const div = document.createElement('div');
    div.innerHTML = ansiToHtml(wrapped);
    this.root.appendChild(div);
  }

  async showAnsiFile(name) {
    const res = await fetch(`/data/ansi/${name}.ans`);
    if (!res.ok) throw new Error(`ansi fetch ${name}: ${res.status}`);
    // Restore real ESC (the files store it as the literal "\e" sequence)
    // and unpack \e[NC cursor-forward escapes into N literal spaces — the
    // menu art uses them for column alignment and ansi_up drops them.
    // \e[2J (clear) and \e[H (home) are no-ops in our scrolling <pre>.
    const text = (await res.text())
      .replace(/\\e/g, '\x1b')
      .replace(/\x1b\[(\d+)C/g, (_, n) => ' '.repeat(parseInt(n, 10)))
      .replace(/\x1b\[2J/g, '')
      .replace(/\x1b\[H/g, '');
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

  // Integer in [min, max]. Empty input + Enter is treated as 0 (the "abort"
  // sentinel in the bank/inn flows). Out-of-range input prints a hint and
  // re-prompts on a new line.
  numbersPrompt(promptText, min, max) {
    return new Promise((resolve) => {
      const attempt = () => {
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

        let buffer = '';
        const onKey = (e) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            document.removeEventListener('keydown', onKey);
            cursor.remove();
            const n = buffer === '' ? 0 : parseInt(buffer, 10);
            if (Number.isNaN(n) || n < min || n > max) {
              this.println(`Please enter a value between ${min} and ${max}.`, 6);
              attempt();
              return;
            }
            resolve(n);
          } else if (e.key === 'Backspace') {
            e.preventDefault();
            buffer = buffer.slice(0, -1);
            inputSpan.textContent = buffer;
          } else if (/^\d$/.test(e.key)) {
            e.preventDefault();
            buffer += e.key;
            inputSpan.textContent = buffer;
          }
        };
        document.addEventListener('keydown', onKey);
      };
      attempt();
    });
  }

  // Mirrors slyio.YesNoQuestion. Resolves true for Y, false for N.
  async yesNoQuestion(prompt) {
    const ch = await this.lettersPrompt(prompt + ' [Y/N]', 'YN');
    return ch === 'Y';
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
