'use strict';

// ── State ─────────────────────────────────────────────
let ws = null;
let username = '';
// Track any keydown handlers registered by prompt builders so we can remove them.
let activeKeyHandlers = [];

// ── DOM refs ──────────────────────────────────────────
const loginScreen    = document.getElementById('login-screen');
const gameScreen     = document.getElementById('game-screen');
const terminalOutput = document.getElementById('terminal-output');
const inputArea      = document.getElementById('input-area');
const statusDot      = document.getElementById('connection-status');
const playerNameEl   = document.getElementById('player-name');
const apiOutputEl    = document.getElementById('api-output');

// ── Login ─────────────────────────────────────────────
document.getElementById('enter-btn').addEventListener('click', startGame);
document.getElementById('username').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') startGame();
});

function startGame() {
  const val = document.getElementById('username').value.trim();
  if (!val) return;
  username = val;
  loginScreen.classList.add('hidden');
  gameScreen.classList.remove('hidden');
  playerNameEl.textContent = username;
  connectWS();
}

// ── WebSocket ──────────────────────────────────────────
function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const url   = `${proto}://${location.host}/ws?username=${encodeURIComponent(username)}`;
  ws = new WebSocket(url);

  ws.addEventListener('open', () => {
    setStatus(true);
    appendOutput('Connected to server.\n', 0);
  });

  ws.addEventListener('message', (evt) => {
    try {
      handleServerMessage(JSON.parse(evt.data));
    } catch (e) {
      console.error('bad ws message', e, evt.data);
    }
  });

  ws.addEventListener('close', () => {
    setStatus(false);
    appendOutput('\n[Disconnected from server]\n', 6);
    clearInputArea();
  });

  ws.addEventListener('error', (e) => {
    console.error('ws error', e);
    setStatus(false);
  });
}

function setStatus(connected) {
  statusDot.classList.toggle('connected',    connected);
  statusDot.classList.toggle('disconnected', !connected);
}

function sendInput(value) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;
  ws.send(JSON.stringify({ type: 'input', value: String(value) }));
}

// ── Server message dispatch ────────────────────────────
function handleServerMessage(msg) {
  switch (msg.type) {
    case 'output':
      appendOutput(msg.text || '', msg.color || 0);
      break;
    case 'clear':
      terminalOutput.innerHTML = '';
      ansiState = { fg: null, bg: null, bold: false };
      break;
    case 'ansi':
      renderAnsiString(msg.text || '');
      break;
    case 'ansi_file':
      renderAnsiString(msg.text || '');
      break;
    case 'error':
      appendOutput((msg.text || 'Server error') + '\n', 6);
      break;
    case 'prompt_letters':
      showLettersPrompt(msg.prompt || '', msg.allowed || '', msg.maxLen || 1);
      break;
    case 'prompt_number':
      showNumberPrompt(msg.prompt || '', msg.min ?? 0, msg.max ?? 9999);
      break;
    case 'prompt_yesno':
      showYesNoPrompt(msg.prompt || '');
      break;
    case 'prompt_pause':
      showPausePrompt(msg.prompt || '--press a key--');
      break;
    case 'prompt_readline':
      showReadlinePrompt(msg.prompt || '', msg.maxLen || 80);
      break;
    default:
      console.warn('unknown msg type', msg.type, msg);
  }
}

// ── ANSI renderer ──────────────────────────────────────
// Standard 16-color ANSI palette
const ANSI_PALETTE = [
  '#000000','#aa0000','#00aa00','#aa5500','#0000aa','#aa00aa','#00aaaa','#aaaaaa',
  '#555555','#ff5555','#55ff55','#ffff55','#5555ff','#ff55ff','#55ffff','#ffffff',
];

let ansiState = { fg: null, bg: null, bold: false };

function renderAnsiString(str) {
  let i = 0, buf = '';
  while (i < str.length) {
    const ch = str[i];
    if (ch === '\x1b' && str[i + 1] === '[') {
      if (buf) { flushAnsiSpan(buf); buf = ''; }
      let j = i + 2;
      while (j < str.length && !/[A-Za-z]/.test(str[j])) j++;
      const cmd = str[j] || '';
      const raw = str.slice(i + 2, j);
      const params = raw.split(';').map(p => (p === '' ? 0 : parseInt(p, 10)));
      applyAnsiEsc(cmd, params);
      i = j + 1;
    } else if (ch === '\r') {
      if (buf) { flushAnsiSpan(buf); buf = ''; }
      // Overwrite current line by clearing it and starting fresh.
      const last = terminalOutput.lastElementChild;
      if (last) last.innerHTML = '';
      i++;
    } else if (ch === '\n') {
      if (buf) { flushAnsiSpan(buf); buf = ''; }
      terminalOutput.appendChild(document.createElement('br'));
      i++;
    } else {
      buf += ch;
      i++;
    }
  }
  if (buf) flushAnsiSpan(buf);
  terminalOutput.scrollTop = terminalOutput.scrollHeight;
}

function flushAnsiSpan(text) {
  const s = document.createElement('span');
  s.textContent = text;
  const css = [];
  if (ansiState.fg) css.push('color:' + ansiState.fg);
  if (ansiState.bg) css.push('background:' + ansiState.bg);
  if (ansiState.bold) css.push('font-weight:bold');
  if (css.length) s.style.cssText = css.join(';');
  terminalOutput.appendChild(s);
}

function applyAnsiEsc(cmd, params) {
  switch (cmd) {
    case 'm': // SGR
      for (const p of (params.length ? params : [0])) {
        if      (p === 0)  { ansiState = { fg: null, bg: null, bold: false }; }
        else if (p === 1)  { ansiState.bold = true; }
        else if (p === 22) { ansiState.bold = false; }
        else if (p >= 30 && p <= 37) {
          ansiState.fg = ANSI_PALETTE[p - 30 + (ansiState.bold ? 8 : 0)];
        }
        else if (p === 39) { ansiState.fg = null; }
        else if (p >= 40 && p <= 47) { ansiState.bg = ANSI_PALETTE[p - 40]; }
        else if (p === 49) { ansiState.bg = null; }
        else if (p >= 90 && p <= 97)  { ansiState.fg = ANSI_PALETTE[p - 90 + 8]; }
        else if (p >= 100 && p <= 107){ ansiState.bg = ANSI_PALETTE[p - 100 + 8]; }
      }
      break;
    case 'J': // Erase display
      if (params[0] === 2 || params[0] === 0) {
        terminalOutput.innerHTML = '';
        ansiState = { fg: null, bg: null, bold: false };
      }
      break;
    case 'K': // Erase line
      { const last = terminalOutput.lastElementChild; if (last) last.innerHTML = ''; }
      break;
    // H/f cursor position and A/B/C/D movement: approximate only
    case 'H': case 'f':
      terminalOutput.appendChild(document.createElement('br'));
      break;
  }
}

// ── Terminal output (plain color-coded) ───────────────
function appendOutput(text, color) {
  const span = document.createElement('span');
  span.className = `c${color}`;
  span.textContent = text;
  terminalOutput.appendChild(span);
  terminalOutput.scrollTop = terminalOutput.scrollHeight;
}

// ── Input area helpers ─────────────────────────────────
function clearInputArea() {
  // Remove all active keyboard handlers first.
  activeKeyHandlers.forEach((h) => document.removeEventListener('keydown', h));
  activeKeyHandlers = [];
  inputArea.innerHTML = '';
}

function addKeyHandler(fn) {
  activeKeyHandlers.push(fn);
  document.addEventListener('keydown', fn);
}

function removeKeyHandler(fn) {
  activeKeyHandlers = activeKeyHandlers.filter((h) => h !== fn);
  document.removeEventListener('keydown', fn);
}

// ── Prompt builders ────────────────────────────────────

function showReadlinePrompt(prompt, maxLen) {
  clearInputArea();

  const container = makeContainer();
  container.appendChild(makeLabel(prompt + ' '));

  const inp = makeTextInput(maxLen || 80);
  const btn = makeBtn('Send', 'choice-btn');

  function submit() {
    const val = inp.value;
    clearInputArea();
    appendOutput(prompt + ' ' + val + '\n', 1);
    sendInput(val);
  }

  btn.addEventListener('click', submit);
  inp.addEventListener('keydown', (e) => { if (e.key === 'Enter') submit(); });

  container.append(inp, btn);
  inputArea.appendChild(container);
  inp.focus();
}

function showLettersPrompt(prompt, allowed, maxLen) {
  clearInputArea();

  const container = makeContainer();
  container.appendChild(makeLabel(prompt + ' '));

  const upper = allowed.toUpperCase();

  if (maxLen === 1 && upper.length > 0 && upper.length <= 12) {
    // Render as clickable letter buttons.
    const btnGroup = document.createElement('div');
    btnGroup.className = 'prompt-btn-group';

    for (const ch of upper) {
      const btn = makeBtn(ch, 'choice-btn');
      btn.addEventListener('click', () => {
        clearInputArea();
        appendOutput(prompt + ' ' + ch + '\n', 1);
        sendInput(ch);
      });
      btnGroup.appendChild(btn);
    }
    container.appendChild(btnGroup);

    // Keyboard shortcut.
    const kh = (e) => {
      const k = e.key.toUpperCase();
      if (upper.includes(k)) {
        removeKeyHandler(kh);
        clearInputArea();
        appendOutput(prompt + ' ' + k + '\n', 1);
        sendInput(k);
      }
    };
    addKeyHandler(kh);
  } else {
    // Multi-char: free-text input.
    const inp = makeTextInput(maxLen || 1);
    const btn = makeBtn('OK', 'choice-btn');

    function submit() {
      const val = inp.value.toUpperCase();
      clearInputArea();
      appendOutput(prompt + ' ' + val + '\n', 1);
      sendInput(val);
    }

    btn.addEventListener('click', submit);
    inp.addEventListener('keydown', (e) => { if (e.key === 'Enter') submit(); });
    container.append(inp, btn);
    inp.focus();
  }

  inputArea.appendChild(container);
}

function showNumberPrompt(prompt, min, max) {
  clearInputArea();

  const container = makeContainer();
  container.appendChild(makeLabel(`${prompt} (${min}-${max}) `));

  const inp = document.createElement('input');
  inp.type = 'number';
  inp.className = 'prompt-input';
  inp.min = min;
  inp.max = max;
  inp.style.maxWidth = '100px';
  inp.autocomplete = 'off';

  const btn = makeBtn('OK', 'choice-btn');

  function submit() {
    const val = parseInt(inp.value, 10);
    if (isNaN(val) || val < min || val > max) {
      inp.style.borderColor = 'var(--red)';
      setTimeout(() => { inp.style.borderColor = ''; }, 800);
      return;
    }
    clearInputArea();
    appendOutput(prompt + ' ' + val + '\n', 1);
    sendInput(String(val));
  }

  btn.addEventListener('click', submit);
  inp.addEventListener('keydown', (e) => { if (e.key === 'Enter') submit(); });

  container.append(inp, btn);
  inputArea.appendChild(container);
  inp.focus();
}

function showYesNoPrompt(prompt) {
  clearInputArea();

  const container = makeContainer();
  container.appendChild(makeLabel(prompt + ' '));

  function choose(val) {
    clearInputArea();
    appendOutput(prompt + ' ' + val + '\n', 1);
    sendInput(val);
  }

  const yBtn = makeBtn('Yes (Y)', 'choice-btn yesno-btn');
  const nBtn = makeBtn('No (N)',  'choice-btn yesno-btn');
  yBtn.addEventListener('click', () => choose('Y'));
  nBtn.addEventListener('click', () => choose('N'));

  const kh = (e) => {
    const k = e.key.toUpperCase();
    if (k === 'Y') { removeKeyHandler(kh); choose('Y'); }
    if (k === 'N') { removeKeyHandler(kh); choose('N'); }
  };
  addKeyHandler(kh);

  container.append(yBtn, nBtn);
  inputArea.appendChild(container);
}

function showPausePrompt(prompt) {
  clearInputArea();

  const container = makeContainer();
  const btn = makeBtn(prompt, 'choice-btn pause-btn');

  function proceed() {
    clearInputArea();
    sendInput('');
  }

  const kh = (e) => {
    if (!e.ctrlKey && !e.altKey && !e.metaKey && e.key !== 'Tab') {
      removeKeyHandler(kh);
      proceed();
    }
  };

  btn.addEventListener('click', proceed);
  addKeyHandler(kh);

  container.appendChild(btn);
  inputArea.appendChild(container);
  btn.focus();
}

// ── DOM helpers ────────────────────────────────────────
function makeContainer() {
  const d = document.createElement('div');
  d.className = 'prompt-container';
  return d;
}

function makeLabel(text) {
  const s = document.createElement('span');
  s.className = 'prompt-label';
  s.textContent = text;
  return s;
}

function makeTextInput(maxLen) {
  const inp = document.createElement('input');
  inp.type = 'text';
  inp.className = 'prompt-input';
  inp.maxLength = maxLen;
  inp.autocomplete = 'off';
  return inp;
}

function makeBtn(label, cls) {
  const btn = document.createElement('button');
  btn.textContent = label;
  btn.className = cls;
  return btn;
}

// ── API Explorer ───────────────────────────────────────
document.querySelectorAll('.api-btn').forEach((btn) => {
  btn.addEventListener('click', async () => {
    const endpoint = btn.dataset.endpoint;
    apiOutputEl.textContent = 'Loading...';
    try {
      const resp = await fetch(endpoint);
      const data = await resp.json();
      apiOutputEl.textContent = JSON.stringify(data, null, 2);
    } catch (e) {
      apiOutputEl.textContent = `Error: ${e.message}`;
    }
  });
});
