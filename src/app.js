const STORAGE_KEY = "raven-chat:e2ee-messages";
const SESSION_KEY = "raven-chat:session";
const CODE_KEY = "raven-chat:email-code";
const DEVICE_SECRET_KEY = "raven-chat:device-secret";
const RETENTION_DAYS = 180;
const encoder = new TextEncoder();
const decoder = new TextDecoder();

const state = {
  session: loadJson(SESSION_KEY, null),
  cryptoKey: null,
  mediaRecorder: null,
  recordedChunks: [],
};

const elements = {
  authForm: document.querySelector("#auth-form"),
  displayName: document.querySelector("#display-name"),
  email: document.querySelector("#email"),
  emailCode: document.querySelector("#email-code"),
  password: document.querySelector("#password"),
  sendCode: document.querySelector("#send-code"),
  messageForm: document.querySelector("#message-form"),
  messageInput: document.querySelector("#message-input"),
  messageList: document.querySelector("#message-list"),
  recordVoice: document.querySelector("#record-voice"),
  exportHtml: document.querySelector("#export-html"),
  statusDot: document.querySelector(".status-dot"),
  toast: document.querySelector("#toast"),
};

init();

async function init() {
  pruneExpiredMessages();
  bindEvents();
  if (state.session) {
    hydrateAuthFields(state.session);
    await unlockSession(state.session);
  }
  await renderMessages();
}

function bindEvents() {
  elements.sendCode.addEventListener("click", () => {
    const code = String(Math.floor(100000 + Math.random() * 900000));
    localStorage.setItem(CODE_KEY, code);
    elements.emailCode.value = code;
    showToast(`Демо-код Raven: ${code}`);
  });

  elements.authForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const displayName = elements.displayName.value.trim();
    const email = elements.email.value.trim().toLowerCase();
    const password = elements.password.value;
    const emailCode = elements.emailCode.value.trim();
    const issuedCode = localStorage.getItem(CODE_KEY);

    if (!displayName || !email) {
      showToast("Укажите имя и почту.");
      return;
    }
    if (!password && (!emailCode || emailCode !== issuedCode)) {
      showToast("Введите пароль или корректный код с почты.");
      return;
    }

    const session = { displayName, email, authMode: password ? "password" : "email-code" };
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    state.session = session;
    await unlockSession(session);
    showToast("Вход выполнен. E2EE-ключ активирован локально.");
  });

  elements.messageForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const text = elements.messageInput.value.trim();
    if (!text) return;
    await saveEncryptedMessage({ type: "text", text });
    elements.messageInput.value = "";
    await renderMessages();
  });

  elements.recordVoice.addEventListener("click", toggleVoiceRecording);
  elements.exportHtml.addEventListener("click", exportConversationHtml);
}

function hydrateAuthFields(session) {
  elements.displayName.value = session.displayName || "";
  elements.email.value = session.email || "";
}

async function unlockSession(session) {
  state.cryptoKey = await deriveKey(`${session.email}:${getDeviceSecret()}`);
  elements.messageInput.disabled = false;
  elements.messageForm.querySelector("button[type='submit']").disabled = false;
  elements.recordVoice.disabled = false;
  elements.statusDot.classList.add("is-online");
}

async function deriveKey(secret) {
  const hash = await crypto.subtle.digest("SHA-256", encoder.encode(secret));
  return crypto.subtle.importKey("raw", hash, { name: "AES-GCM" }, false, ["encrypt", "decrypt"]);
}

async function saveEncryptedMessage(payload) {
  if (!state.session || !state.cryptoKey) {
    showToast("Сначала войдите в Raven Chat.");
    return;
  }
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const plaintext = JSON.stringify(payload);
  const ciphertext = await crypto.subtle.encrypt({ name: "AES-GCM", iv }, state.cryptoKey, encoder.encode(plaintext));
  const messages = getStoredMessages();
  messages.push({
    id: crypto.randomUUID(),
    author: state.session.displayName,
    email: state.session.email,
    createdAt: new Date().toISOString(),
    expiresAt: new Date(Date.now() + RETENTION_DAYS * 24 * 60 * 60 * 1000).toISOString(),
    iv: toBase64(iv),
    ciphertext: toBase64(new Uint8Array(ciphertext)),
  });
  localStorage.setItem(STORAGE_KEY, JSON.stringify(messages));
}

async function decryptMessage(message) {
  try {
    const decrypted = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv: fromBase64(message.iv) },
      state.cryptoKey,
      fromBase64(message.ciphertext),
    );
    return JSON.parse(decoder.decode(decrypted));
  } catch {
    return { type: "locked", text: "Не удалось расшифровать сообщение этим локальным ключом." };
  }
}

async function renderMessages() {
  const messages = getStoredMessages();
  elements.messageList.innerHTML = "";
  if (!messages.length) {
    elements.messageList.innerHTML = `
      <div class="empty-state">
        <div class="empty-state__icon" aria-hidden="true">☾</div>
        <h3>Темная комната Raven готова</h3>
        <p>Войдите по почте, коду или паролю и отправьте первое зашифрованное сообщение.</p>
      </div>`;
    return;
  }

  for (const message of messages) {
    const payload = state.cryptoKey ? await decryptMessage(message) : { type: "locked", text: "Сообщение зашифровано." };
    elements.messageList.appendChild(createMessageNode(message, payload));
  }
  elements.messageList.scrollTop = elements.messageList.scrollHeight;
}

function createMessageNode(message, payload) {
  const article = document.createElement("article");
  article.className = `message ${state.session?.email === message.email ? "is-own" : ""}`;

  const meta = document.createElement("div");
  meta.className = "message__meta";
  meta.textContent = `${message.author} · ${new Date(message.createdAt).toLocaleString("ru-RU")} · E2EE`;
  article.appendChild(meta);

  if (payload.type === "voice") {
    const label = document.createElement("p");
    label.className = "message__text";
    label.textContent = "Голосовое сообщение";
    const audio = document.createElement("audio");
    audio.controls = true;
    audio.src = payload.dataUrl;
    article.append(label, audio);
  } else {
    const text = document.createElement("p");
    text.className = "message__text";
    text.textContent = payload.text;
    article.appendChild(text);
  }
  return article;
}

async function toggleVoiceRecording() {
  if (state.mediaRecorder?.state === "recording") {
    state.mediaRecorder.stop();
    return;
  }
  if (!navigator.mediaDevices?.getUserMedia) {
    showToast("Запись голоса недоступна в этом браузере.");
    return;
  }
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
    state.recordedChunks = [];
    state.mediaRecorder = new MediaRecorder(stream);
    state.mediaRecorder.addEventListener("dataavailable", (event) => {
      if (event.data.size) state.recordedChunks.push(event.data);
    });
    state.mediaRecorder.addEventListener("stop", async () => {
      stream.getTracks().forEach((track) => track.stop());
      elements.recordVoice.classList.remove("is-recording");
      elements.recordVoice.textContent = "Голосовое";
      const blob = new Blob(state.recordedChunks, { type: state.mediaRecorder.mimeType || "audio/webm" });
      const dataUrl = await blobToDataUrl(blob);
      await saveEncryptedMessage({ type: "voice", dataUrl, mimeType: blob.type });
      await renderMessages();
      showToast("Голосовое сообщение зашифровано и сохранено.");
    });
    state.mediaRecorder.start();
    elements.recordVoice.classList.add("is-recording");
    elements.recordVoice.textContent = "Остановить";
  } catch {
    showToast("Не удалось получить доступ к микрофону.");
  }
}

async function exportConversationHtml() {
  const messages = getStoredMessages();
  const rows = [];
  for (const message of messages) {
    const payload = state.cryptoKey ? await decryptMessage(message) : { type: "locked", text: "Сообщение зашифровано." };
    const content = payload.type === "voice"
      ? `<audio controls src="${escapeAttribute(payload.dataUrl)}"></audio>`
      : `<p>${escapeHtml(payload.text)}</p>`;
    rows.push(`<article><h2>${escapeHtml(message.author)}</h2><time>${escapeHtml(new Date(message.createdAt).toLocaleString("ru-RU"))}</time>${content}</article>`);
  }
  const html = `<!doctype html><html lang="ru"><head><meta charset="utf-8"><title>Raven Chat export</title><style>body{font-family:system-ui;background:#06070b;color:#edf2ff;padding:32px}article{border:1px solid #263244;border-radius:18px;margin:0 0 16px;padding:16px;background:#0f172a}time{color:#94a3b8}audio{display:block;margin-top:12px;max-width:100%}</style></head><body><h1>Raven Chat export</h1><p>Экспорт E2EE-переписки, расшифрованный локально в браузере.</p>${rows.join("")}</body></html>`;
  downloadFile(`raven-chat-export-${new Date().toISOString().slice(0, 10)}.html`, html, "text/html");
  showToast("HTML-экспорт подготовлен.");
}

function pruneExpiredMessages() {
  const now = Date.now();
  const active = getStoredMessages().filter((message) => new Date(message.expiresAt).getTime() > now);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(active));
}

function getStoredMessages() { return loadJson(STORAGE_KEY, []); }
function getDeviceSecret() {
  let secret = localStorage.getItem(DEVICE_SECRET_KEY);
  if (!secret) {
    secret = crypto.randomUUID();
    localStorage.setItem(DEVICE_SECRET_KEY, secret);
  }
  return secret;
}
function loadJson(key, fallback) {
  try { return JSON.parse(localStorage.getItem(key)) ?? fallback; } catch { return fallback; }
}
function toBase64(bytes) { return btoa(String.fromCharCode(...bytes)); }
function fromBase64(value) { return Uint8Array.from(atob(value), (char) => char.charCodeAt(0)); }
function blobToDataUrl(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}
function downloadFile(filename, content, type) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}
function escapeHtml(value = "") {
  return String(value).replace(/[&<>'"]/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[char]);
}
function escapeAttribute(value = "") { return escapeHtml(value).replace(/`/g, "&#96;"); }
function showToast(message) {
  elements.toast.textContent = message;
  elements.toast.classList.add("is-visible");
  clearTimeout(showToast.timeout);
  showToast.timeout = setTimeout(() => elements.toast.classList.remove("is-visible"), 3200);
}
