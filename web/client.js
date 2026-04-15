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
      break;
    case 'ansi':
      // Raw ANSI codes are no-ops in the browser.
      break;
    case 'ansi_file':
      appendOutput(`[ANSI art: ${msg.text}]\n`, 0);
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

// ── Terminal output ────────────────────────────────────
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
