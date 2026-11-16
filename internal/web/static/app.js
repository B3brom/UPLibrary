const API = '';
let token = localStorage.getItem('token') || '';
let currentUser = null;
try { currentUser = JSON.parse(localStorage.getItem('user') || 'null'); } catch { currentUser = null; }

if (token && currentUser && (typeof currentUser.role_id !== 'string' || !currentUser.role_id.includes('-'))) {
    token = ''; currentUser = null;
    localStorage.removeItem('token'); localStorage.removeItem('user');
}
if (token && !currentUser) { token = ''; localStorage.removeItem('token'); }

const ADMIN_ROLE_ID = '00000000-0000-0000-0000-000000000004';
const MANAGER_ROLE_ID = '00000000-0000-0000-0000-000000000003';
const CLIENT_CANCEL_REASONS = [
    'Передумал(а) оформлять заказ',
    'Ошибся(лась) с размером',
    'Ошибся(лась) с цветом',
    'Нашёл(ла) дешевле',
    'Долгая доставка',
];

const MANAGER_CANCEL_REASONS = [
    'Клиент не отвечает',
    'Товар закончился на складе',
    'Брак / повреждение товара',
    'Клиент попросил отменить заказ',
    'Возврат по требованию клиента',
    'Проблема с оплатой',
];
function isAdmin() { return currentUser && currentUser.role_id === ADMIN_ROLE_ID; }
function isManager() { return currentUser && currentUser.role_id === MANAGER_ROLE_ID; }
function canManageProducts() { return isAdmin() || isManager(); }

// ---------- utils ----------
function toast(msg, isError = false) {
    const el = document.getElementById('toast');
    el.textContent = msg;
    el.className = 'toast show' + (isError ? ' error' : '');
    setTimeout(() => el.className = 'toast', 2500);
}

async function api(path, opts = {}) {
    const headers = { ...(opts.headers || {}) };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    if (opts.body && !(opts.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
    }
    const res = await fetch(API + path, { ...opts, headers });
    if (!res.ok) {
        const text = await res.text();
        let message = text || res.statusText;
        try { message = JSON.parse(text).error || message; } catch { }
        if (res.status === 401 && token && path !== '/login' && path !== '/register') {
            clearSession();
        }
        throw new Error(message);
    }
    const ct = res.headers.get('content-type') || '';
    return ct.includes('json') ? res.json() : res.text();
}

function escapeHtml(s) {
    return String(s ?? '').replace(/[&<>"']/g, c => ({
        '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
    })[c]);
}

function shortId(id) { return id ? id.slice(0, 8) : ''; }

// ---------- session ----------
function saveSession(session) {
    token = session.token;
    currentUser = session.user;
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(currentUser));
    updateHeaderAuth();
    updateTabs();
    updateAdminTab()
    renderBalance();
    closeAuth();
}

function clearSession() {
    token = ''; currentUser = null;
    localStorage.removeItem('token'); localStorage.removeItem('user');
    updateHeaderAuth();
    updateTabs()
    updateAdminTab()
    renderBalance();
}

function updateHeaderAuth() {
    const auth = document.getElementById('open-auth-btn');
    const logout = document.getElementById('logout');
    const who = document.getElementById('whoami');
    if (currentUser) {
        auth.style.display = 'none';
        logout.style.display = 'inline-block';
        who.textContent = currentUser.name;
    } else {
        auth.style.display = 'inline-block';
        logout.style.display = 'none';
        who.textContent = '';
    }
}

function requireAuth() {
    if (token && currentUser) return true;
    toast('Сначала войдите или зарегистрируйтесь', true);
    openAuth();
    return false;
}

// ---------- auth modal ----------
function openAuth() {
    document.getElementById('auth-modal').style.display = 'flex';
    showLoginForm();
}
function closeAuth() {
    document.getElementById('auth-modal').style.display = 'none';
}
function showLoginForm() {
    document.getElementById('login-form').style.display = 'flex';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('show-login').classList.add('active');
    document.getElementById('show-register').classList.remove('active');
}
function showRegisterForm() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'flex';
    document.getElementById('show-register').classList.add('active');
    document.getElementById('show-login').classList.remove('active');
}

document.getElementById('open-auth-btn').onclick = openAuth;
document.getElementById('show-login').onclick = showLoginForm;
document.getElementById('show-register').onclick = showRegisterForm;

document.getElementById('login').onclick = async (e) => {
    e.preventDefault();
    try {
        const r = await api('/login', {
            method: 'POST',
            body: JSON.stringify({
                email: document.getElementById('login-email').value,
                password: document.getElementById('login-password').value,
            }),
        });
        saveSession(r);
        toast('Вход выполнен');
    } catch (e) { toast('Ошибка входа: ' + e.message, true); }
};

document.getElementById('register').onclick = async (e) => {
    e.preventDefault();
    if (!document.getElementById('reg-consent').checked) {
        toast('Нужно согласие на обработку персональных данных', true);
        return;
    }
    try {
        const r = await api('/register', {
            method: 'POST',
            body: JSON.stringify({
                name: document.getElementById('reg-name').value,
                email: document.getElementById('reg-email').value,
                phone: document.getElementById('reg-phone').value,
                password: document.getElementById('reg-password').value,
            }),
        });
        saveSession(r);
        toast('Аккаунт создан');
    } catch (e) { toast('Ошибка регистрации: ' + e.message, true); }
};

document.getElementById('logout').onclick = () => {
    clearSession();
    toast('Вы вышли');
};

// ---------- balance ----------
function renderBalance() {
    const el = document.getElementById('balance-display');
    if (!el) return;
    if (currentUser) {
        document.getElementById('balance-value').textContent = (currentUser.balance ?? 0).toFixed(2);
        el.style.display = 'inline-flex';
        el.style.alignItems = 'center';
        el.style.gap = '6px';
    } else {
        el.style.display = 'none';
    }
}

async function refreshMe() {
    if (!currentUser) return;
    try {
        const u = await api('/me');
        currentUser.balance = u.balance;
        currentUser.name = u.name;
        currentUser.role_id = u.role_id;
        localStorage.setItem('user', JSON.stringify(currentUser));
        document.getElementById('whoami').textContent = currentUser.name;
        renderBalance();
        updateTabs()
    } catch { }
}

function openTopup() {
    document.getElementById('topup-current').textContent = (currentUser?.balance ?? 0).toFixed(2);
    document.getElementById('topup-amount').value = '';
    document.getElementById('topup-modal').style.display = 'flex';
}
function closeTopup() {
    document.getElementById('topup-modal').style.display = 'none';
}
async function submitTopup() {
    const amount = parseFloat(document.getElementById('topup-amount').value);
    if (!amount || amount <= 0) { toast('Введите сумму > 0', true); return; }
    try {
        const r = await api('/balance/topup', { method: 'POST', body: JSON.stringify({ amount }) });
        currentUser.balance = r.balance;
        localStorage.setItem('user', JSON.stringify(currentUser));
        renderBalance();
        closeTopup();
        toast('Баланс пополнен');
    } catch (e) { toast(e.message, true); }
}
document.getElementById('topup-btn').onclick = openTopup;

// ---------- tabs ----------
function updateTabs() {
    const role = currentUser ? currentUser.role_id : null;
    const isStaff = role === ADMIN_ROLE_ID || role === MANAGER_ROLE_ID;

    const cartTab  = document.querySelector('.tab[data-tab="cart"]');
    const adminTab = document.getElementById('admin-tab');

    // Управление/Админ — только менеджеру и админу
    if (adminTab) {
        adminTab.style.display = isStaff ? 'inline-block' : 'none';
        adminTab.textContent = role === ADMIN_ROLE_ID ? 'Админ' : 'Управление';
    }

    // Если активная вкладка стала невидимой — уходим в каталог
    const active = document.querySelector('.tab.active');
    if (active && active.style.display === 'none') {
        const catalogTab = document.querySelector('.tab[data-tab="catalog"]');
        if (catalogTab) catalogTab.click();
    }
}

function updateAdminTab() {
    const adminBlock = document.getElementById('admin-only-block');
    const usersBtn = document.getElementById('open-users-btn');
    if (adminBlock) adminBlock.style.display = isAdmin() ? 'block' : 'none';
    if (usersBtn) usersBtn.style.display = isAdmin() ? 'inline-block' : 'none';

    const usersModal = document.getElementById('users-modal');
    if (usersModal && !isAdmin() && usersModal.style.display === 'flex') {
        usersModal.style.display = 'none';
    }
    updateTabs();
}
document.querySelectorAll('.tab').forEach(btn => {
    btn.onclick = () => {
        if (['cart', 'orders'].includes(btn.dataset.tab) && !requireAuth()) return;
        document.querySelectorAll('.tab').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
        btn.classList.add('active');
        document.getElementById(btn.dataset.tab).classList.add('active');
        if (btn.dataset.tab === 'cart') loadCart();
        if (btn.dataset.tab === 'orders') loadOrders();
    };
});

// ---------- catalog ----------
async function loadProducts() {
    const fd = new FormData(document.getElementById('filters'));
    const params = new URLSearchParams();
    for (const [k, v] of fd) if (v) params.append(k, v);
    try {
        const list = await api('/product?' + params.toString());
        renderProducts(list || []);
    } catch (e) { toast(e.message, true); }
}

function renderProducts(list) {
    const root = document.getElementById('products');
    if (!list.length) { root.innerHTML = '<p class="hint">Товаров нет</p>'; return; }
    const canManage = canManageProducts();
    root.innerHTML = list.map(p => `
    <div class="card" onclick="openProductCard('${p.id}')">
      <div class="card-img-wrap">
        ${p.main_photo_id
            ? `<img class="card-img" src="/photo/${p.main_photo_id}" alt="${escapeHtml(p.name)}">`
            : `<div class="card-img placeholder">нет фото</div>`}
      </div>
      <h3>${escapeHtml(p.name)}</h3>
      <div class="meta">${escapeHtml(p.brand || '')} · ${escapeHtml(p.type || '')} · ${escapeHtml(p.gender || '')}/${escapeHtml(p.size || '')} · ${escapeHtml(p.season || '')}</div>
      <div class="price">${p.price} ₽</div>
      <div class="meta">В наличии: <b>${p.quantity ?? 0}</b></div>
      <div class="meta">${escapeHtml(p.description || '')}</div>
      <div class="actions" onclick="event.stopPropagation()">
        <button onclick="addToCart('${p.id}')" ${(p.quantity ?? 0) <= 0 ? 'disabled' : ''}>В корзину</button>
        ${canManage ? `
          <button onclick="startEditProduct('${p.id}')">Изменить</button>
          <button onclick="deleteProduct('${p.id}')">Удалить</button>
        ` : ''}
      </div>
    </div>`).join('');
}

document.getElementById('filters').onsubmit = e => { e.preventDefault(); loadProducts(); };
document.getElementById('reset-filters').onclick = () => {
    document.getElementById('filters').reset();
    loadProducts();
};

// ---------- product card modal ----------
async function openProductCard(id) {
    try {
        const [p, photos] = await Promise.all([
            api('/product/' + id),
            api('/photo?product_id=' + id).catch(() => []),
        ]);

        const mainPhoto = photos[0];
        const mainHtml = mainPhoto
            ? `<img src="/photo/${mainPhoto.id}" class="pc-main-img"
                    onclick="openGallery('${id}', 0)" alt="${escapeHtml(p.name)}">`
            : `<div class="pc-main-placeholder">нет фото</div>`;

        const thumbsHtml = photos.length > 1
            ? `<div class="pc-thumbs">${photos.map((ph, i) =>
                `<img src="/photo/${ph.id}" class="pc-thumb${i === 0 ? ' active' : ''}"
                      onclick="openGallery('${id}', ${i})" alt="">`).join('')}</div>`
            : '';

        document.getElementById('product-card-body').innerHTML = `
          <div class="pc-layout">
            <div class="pc-info">
              <h2>${escapeHtml(p.name)}</h2>
              <div class="pc-price">${p.price} ₽</div>
              <div class="pc-meta-grid">
                <div><b>Тип:</b> ${escapeHtml(p.type || '—')}</div>
                <div><b>Пол:</b> ${escapeHtml(p.gender || '—')}</div>
                <div><b>Размер:</b> ${escapeHtml(p.size || '—')}</div>
                <div><b>Цвет:</b> ${escapeHtml(p.color || '—')}</div>
                <div><b>Бренд:</b> ${escapeHtml(p.brand || '—')}</div>
                <div><b>Сезон:</b> ${escapeHtml(p.season || '—')}</div>
                <div><b>В наличии:</b> ${p.quantity ?? 0}</div>
              </div>
              <p class="pc-desc">${escapeHtml(p.description || '')}</p>
              <div class="pc-actions">
                <button onclick="addToCart('${p.id}')" ${(p.quantity ?? 0) <= 0 ? 'disabled' : ''}>
                  В корзину
                </button>
              </div>
            </div>
            <div class="pc-media">
              ${mainHtml}
              ${thumbsHtml}
            </div>
          </div>
        `;
        document.getElementById('product-modal').style.display = 'flex';
    } catch (e) { toast(e.message, true); }
}
function closeProductCard() {
    document.getElementById('product-modal').style.display = 'none';
}

// ---------- gallery ----------
let galleryState = { photos: [], index: 0 };

async function openGallery(productId, startIndex = 0) {
    try {
        const photos = await api('/photo?product_id=' + productId);
        if (!photos || !photos.length) { toast('У товара нет фотографий', true); return; }
        galleryState.photos = photos;
        galleryState.index = Math.max(0, Math.min(startIndex, photos.length - 1));
        renderGallery();
        document.getElementById('gallery').style.display = 'flex';
    } catch (e) { toast(e.message, true); }
}

function renderGallery() {
    const { photos, index } = galleryState;
    const body = document.getElementById('gallery-body');
    const many = photos.length > 1;
    body.innerHTML = `
      <div class="lightbox">
        <img src="/photo/${photos[index].id}" class="lb-img" alt="">
        ${many ? `<button class="lb-nav lb-prev" onclick="galleryPrev(event)">‹</button>` : ''}
        ${many ? `<button class="lb-nav lb-next" onclick="galleryNext(event)">›</button>` : ''}
      </div>
      <div class="lb-counter">${index + 1} / ${photos.length}</div>
    `;
}

function galleryPrev(e) {
    if (e) e.stopPropagation();
    const n = galleryState.photos.length;
    if (n < 2) return;
    galleryState.index = (galleryState.index - 1 + n) % n;
    renderGallery();
}

function galleryNext(e) {
    if (e) e.stopPropagation();
    const n = galleryState.photos.length;
    if (n < 2) return;
    galleryState.index = (galleryState.index + 1) % n;
    renderGallery();
}

function closeGallery() {
    document.getElementById('gallery').style.display = 'none';
}

document.addEventListener('keydown', e => {
    if (document.getElementById('gallery').style.display !== 'flex') return;
    if (e.key === 'ArrowLeft') galleryPrev();
    if (e.key === 'ArrowRight') galleryNext();
    if (e.key === 'Escape') closeGallery();
});

// ---------- cart ----------
async function addToCart(productId) {
    if (!requireAuth()) return;
    try {
        await api('/cart', {
            method: 'POST',
            body: JSON.stringify({ product_id: productId, count: 1 }),
        });
        toast('Добавлено');
    } catch (e) { toast(e.message, true); }
}

function pluralizeItems(n) {
    const mod10 = n % 10;
    const mod100 = n % 100;
    if (mod10 === 1 && mod100 !== 11) return 'товар';
    if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return 'товара';
    return 'товаров';
}

async function loadCart() {
    const hint = document.getElementById('cart-hint');
    const root = document.getElementById('cart-items');
    if (!requireAuth()) {
        hint.textContent = 'Войдите или зарегистрируйтесь, чтобы пользоваться корзиной.';
        hint.style.display = 'block';
        root.innerHTML = '';
        return;
    }
    hint.style.display = 'none';
    await refreshMe();
    try {
        const items = await api('/cart/' + currentUser.id);
        if (!items || !items.length) {
            root.innerHTML = '<p class="hint">Корзина пуста</p>';
            return;
        }
        const enriched = await Promise.all(items.map(async c => {
            try {
                const p = await api('/product/' + c.product_id);
                return { ...c, name: p.name, price: p.price, stock: p.quantity, main_photo_id: p.main_photo_id };
            } catch {
                return { ...c, name: '⚠ товар удалён', price: 0, stock: 0, main_photo_id: null };
            }
        }));

        const totalSum = enriched.reduce((s, c) => s + c.price * c.count, 0);
        const totalCount = enriched.reduce((s, c) => s + c.count, 0);
        const itemsWord = pluralizeItems(totalCount);
        const notEnough = totalSum > (currentUser.balance ?? 0);

        root.innerHTML = enriched.map(c => `
          <div class="row cart-row">
            ${c.main_photo_id
                ? `<img src="/photo/${c.main_photo_id}" class="cart-thumb">`
                : `<div class="cart-thumb placeholder">нет фото</div>`}
            <span class="grow">${escapeHtml(c.name)}</span>
            <span>${c.price.toFixed(2)} ₽</span>
            <span class="cart-controls">
              <button onclick="changeQty('${c.id}', ${c.count - 1})">−</button>
              <input type="number" class="cart-qty-input" min="1" value="${c.count}"
                     onchange="changeQty('${c.id}', parseInt(this.value) || 1)">
              <button onclick="changeQty('${c.id}', ${c.count + 1})">+</button>
            </span>
            <span>${(c.price * c.count).toFixed(2)} ₽</span>
            <button onclick="removeCart('${c.id}')">Удалить</button>
          </div>
        `).join('') + `
          <div class="cart-total">
            <div class="cart-total-line">
              Всего: <b>${totalCount}</b> ${itemsWord}
            </div>
            <div class="cart-total-line">
              Итого: <b>${totalSum.toFixed(2)} ₽</b>
              ${notEnough
                ? `<span class="hint" style="color:#c0392b; margin-left:8px">
                     (не хватает ${(totalSum - (currentUser.balance ?? 0)).toFixed(2)} ₽)
                   </span>`
                : ''}
            </div>
          </div>`;
    } catch (e) { toast(e.message, true); }
}

async function changeQty(id, newCount) {
    if (newCount < 1) return removeCart(id);
    try {
        await api('/cart/' + id, {
            method: 'PUT',
            body: JSON.stringify({ count: newCount }),
        });
        loadCart();
    } catch (e) { toast(e.message, true); }
}

async function removeCart(id) {
    try {
        await api('/cart/' + id, { method: 'DELETE' });
        loadCart();
    } catch (e) { toast(e.message, true); }
}

document.getElementById('checkout').onclick = async () => {
    if (!requireAuth()) return;
    try {
        const cartItems = await api('/cart/' + currentUser.id);
        if (!cartItems || !cartItems.length) { toast('Корзина пуста', true); return; }
        const orderItems = cartItems.filter(c => c.product_id)
            .map(c => ({ product_id: c.product_id, quantity: c.count }));
        if (!orderItems.length) { toast('В корзине нет валидных товаров', true); return; }
        const o = await api('/order', {
            method: 'POST',
            body: JSON.stringify({ items: orderItems }),
        });
        for (const c of cartItems) {
            try { await api('/cart/' + c.id, { method: 'DELETE' }); } catch { }
        }
        await refreshMe();
        toast('Заказ №' + shortId(o.id) + ' оформлен');
        loadCart();
    } catch (e) { toast(e.message, true); }
};

// ---------- orders ----------
async function loadOrders() {
    if (!requireAuth()) return;
    try {
        const canManageOrders = canManageProducts();
        const items = await api(canManageOrders ? '/order?limit=50' : '/my/orders?limit=50');
        const root = document.getElementById('orders-list');
        if (!items || !items.length) {
            root.innerHTML = '<p class="hint">Заказов нет</p>';
            return;
        }

        const canChange = ['Принят', 'В работе', 'Собран', 'Готов к выдаче'];

        root.innerHTML = items.map(o => {
            const cancelled = o.status === 'Отменён' || o.refunded;
            const issued    = o.status === 'Получен';
            const locked    = cancelled || issued;

            const clientCanCancel  = !canManageOrders && !locked && o.status === 'Принят';
            const managerCanCancel =  canManageOrders && !locked &&
                ['Принят', 'В работе', 'Собран', 'Готов к выдаче'].includes(o.status);
            const canCancel = clientCanCancel || managerCanCancel;

            const canIssue = !locked && o.status === 'Готов к выдаче' && canManageOrders;

            // Явное объявление — источник прошлой ошибки был здесь
            let statusLabel;
            if (cancelled)      statusLabel = '⛔ Отменён';
            else if (issued)    statusLabel = '✅ Получен';
            else                statusLabel = o.status;

            return `
              <div class="order-row" data-order-id="${o.id}">
                <div class="order-head">
                  <span class="grow">
                    №${shortId(o.id)} · <b>${escapeHtml(o.user_name)}</b>
                    · ${o.items_count} вещ. · ${o.total.toFixed(2)} ₽ · ${statusLabel}
                  </span>
                  <button class="toggle-details" data-open="0">Подробнее</button>
                  ${canIssue ? `<button class="issue-order" onclick="issueOrder('${o.id}')">Выдать</button>` : ''}
                  ${canCancel ? `<button class="cancel-order" onclick="openCancelModal('${o.id}')">Отменить</button>` : ''}
                  ${canManageOrders && !locked
                    ? `<select class="status-select" onchange="setStatus('${o.id}', this.value, this)">
                         ${canChange.map(s =>
                            `<option ${s === o.status ? 'selected' : ''}>${s}</option>`).join('')}
                       </select>`
                    : ''}
                </div>
                ${cancelled && o.cancel_reason
                    ? `<div class="cancel-reason-line">Причина отмены: ${escapeHtml(o.cancel_reason)}</div>`
                    : ''}
                <div class="order-details" style="display:none"></div>
              </div>`;
        }).join('');

        root.querySelectorAll('.toggle-details').forEach(btn => {
            btn.onclick = () => toggleOrderDetails(btn);
        });
    } catch (e) {
        toast(e.message, true);
        console.error('loadOrders:', e);
    }
}
async function toggleOrderDetails(btn) {
    const row = btn.closest('.order-row');
    const details = row.querySelector('.order-details');
    const orderId = row.dataset.orderId;
    if (btn.dataset.open === '1') {
        details.style.display = 'none';
        btn.dataset.open = '0';
        btn.textContent = 'Подробнее';
        return;
    }
    try {
        const d = await api('/order/' + orderId);
        const itemsHtml = (d.items || []).map(it => `
          <div class="order-item">
            ${it.photo_id
                ? `<img src="/photo/${it.photo_id}" class="order-item-img" onclick="openGallery('${it.product_id}')">`
                : `<div class="order-item-img placeholder">нет фото</div>`}
            <div class="grow">
              <div>${escapeHtml(it.product_name)} × ${it.quantity}</div>
              <div class="uuid-line">Артикул товара: <code>${it.product_id}</code></div>
            </div>
            <span>${(it.price * it.quantity).toFixed(2)} ₽</span>
          </div>
        `).join('');
        const cancelInfo = d.cancel_reason
            ? `<div class="cancel-reason-block">Причина отмены: ${escapeHtml(d.cancel_reason)}</div>`
            : '';
        details.innerHTML = `
          ${cancelInfo}
          <div class="order-uuid-block">
            <div>Артикул заказа: <code>${d.id}</code></div>
            <div>UUID пользователя: <code>${d.user_id}</code></div>
            <div>Дата: ${new Date(d.datetime).toLocaleString('ru-RU')}</div>
          </div>
          ${itemsHtml || '<p class="hint">Нет позиций</p>'}
          <div style="text-align:right; padding:8px 0; font-weight:600">Итого: ${(d.total ?? 0).toFixed(2)} ₽</div>
        `;
        details.style.display = 'block';
        btn.dataset.open = '1';
        btn.textContent = 'Скрыть';
    } catch (e) { toast(e.message, true); }
}

let cancelOrderId = null;

function openCancelModal(orderId) {
    cancelOrderId = orderId;
    const isStaff = canManageProducts();
    const reasons = isStaff ? MANAGER_CANCEL_REASONS : CLIENT_CANCEL_REASONS;

    // Контейнер: ищем по id, затем по классу, при отсутствии — создаём
    let container = document.getElementById('cancel-reasons-list');
    if (!container) {
        container = document.querySelector('#cancel-modal .cancel-reasons');
    }
    if (!container) {
        const modal = document.getElementById('cancel-modal');
        const content = modal && modal.querySelector('.modal-content');
        const h3 = content && content.querySelector('h3');
        if (!h3) {
            console.error('[cancel] не найдена модалка #cancel-modal или её .modal-content/h3');
            return;
        }
        container = document.createElement('div');
        container.className = 'cancel-reasons';
        container.id = 'cancel-reasons-list';
        h3.insertAdjacentElement('afterend', container);
    }
    // Нормализуем id и класс — на случай старой разметки
    container.id = 'cancel-reasons-list';
    if (!container.classList.contains('cancel-reasons')) {
        container.classList.add('cancel-reasons');
    }

    container.innerHTML = reasons.map(r =>
        `<label><input type="radio" name="cancel-reason" value="${escapeHtml(r)}"> ${escapeHtml(r)}</label>`
    ).join('') + `
        <label><input type="radio" name="cancel-reason" value="__other__"> Другое:</label>
        <input type="text" id="cancel-reason-custom" placeholder="Своя причина"
               maxlength="255" style="display:none">
    `;

    container.querySelectorAll('input[name="cancel-reason"]').forEach(el => {
        el.onchange = () => {
            const custom = document.getElementById('cancel-reason-custom');
            custom.style.display = el.value === '__other__' ? 'block' : 'none';
            if (el.value !== '__other__') custom.value = '';
        };
    });

    document.getElementById('cancel-modal').style.display = 'flex';
}

function closeCancelModal() {
    cancelOrderId = null;
    document.getElementById('cancel-modal').style.display = 'none';
}

async function submitCancel() {
    if (!cancelOrderId) { toast('Заказ не выбран', true); return; }
    const sel = document.querySelector('input[name="cancel-reason"]:checked');
    if (!sel) { toast('Выберите причину отмены', true); return; }
    let reason = sel.value;
    if (reason === '__other__') {
        reason = document.getElementById('cancel-reason-custom').value.trim();
        if (!reason) { toast('Опишите причину', true); return; }
    }
    try {
        await api('/order/' + cancelOrderId + '/cancel', {
            method: 'POST',
            body: JSON.stringify({ reason }),
        });
        toast('Заказ отменён');
        closeCancelModal();
        await refreshMe();
        loadOrders();
    } catch (e) { toast(e.message, true); }
}

async function issueOrder(id) {
    if (!confirm('Отметить заказ как полученный? После этого изменить его будет нельзя.')) return;
    try {
        await api('/order/' + id, {
            method: 'PUT',
            body: JSON.stringify({ status: 'Получен' }),
        });
        toast('Заказ выдан клиенту');
        loadOrders();
    } catch (e) { toast(e.message, true); }
}

async function setStatus(id, status, selectEl) {
    try {
        await api('/order/' + id, { method: 'PUT', body: JSON.stringify({ status }) });
        toast('Статус обновлён');
    } catch (e) {
        toast(e.message, true);
        if (selectEl) loadOrders();
    }
}

// ---------- product form ----------
function resetProductForm() {
    const form = document.getElementById('add-product');
    form.reset();
    form.querySelector('[name=id]').value = '';
    form.querySelector('[name=quantity]').value = 0;
    document.getElementById('product-submit').textContent = 'Создать';
    document.getElementById('product-cancel').style.display = 'none';
    document.getElementById('product-form-title').textContent = 'Добавить товар';
    document.getElementById('product-photos-block').style.display = 'none';
}

document.getElementById('product-cancel').onclick = resetProductForm;

document.getElementById('add-product').onsubmit = async (e) => {
    e.preventDefault();
    const form = e.target;
    const data = Object.fromEntries(new FormData(form));
    const id = data.id;
    delete data.id;
    data.price = parseFloat(data.price);
    data.quantity = parseInt(data.quantity) || 0;

    const photoInput = document.getElementById('new-product-photo');
    const photoFiles = photoInput.files ? Array.from(photoInput.files) : [];

    try {
        let productId = id;
        if (id) {
            await api('/product/' + id, { method: 'PUT', body: JSON.stringify(data) });
            toast('Товар обновлён');
        } else {
            const r = await api('/product', { method: 'POST', body: JSON.stringify(data) });
            productId = r.product.id;
            if (r.restocked) toast('Товар уже есть, количество увеличено');
            else toast('Создан: ' + shortId(productId));
        }
        for (const file of photoFiles) {
            const fd = new FormData();
            fd.append('file', file);
            fd.append('product_id', productId);
            try { await api('/photo', { method: 'POST', body: fd }); }
            catch (err) { toast('Фото: ' + err.message, true); }
        }
        loadProducts();
        await startEditProduct(productId);
    } catch (err) { toast(err.message, true); }
};

async function startEditProduct(id) {
    try {
        const p = await api('/product/' + id);
        const form = document.getElementById('add-product');
        for (const f of ['name', 'type', 'gender', 'size', 'color', 'brand', 'season', 'description']) {
            const el = form.querySelector('[name=' + f + ']');
            if (el) el.value = p[f] ?? '';
        }
        form.querySelector('[name=price]').value = p.price;
        form.querySelector('[name=quantity]').value = p.quantity ?? 0;
        form.querySelector('[name=id]').value = p.id;
        document.getElementById('product-submit').textContent = 'Сохранить';
        document.getElementById('product-cancel').style.display = 'inline-block';
        document.getElementById('product-form-title').textContent = 'Редактирование: ' + p.name;
        form.scrollIntoView({ behavior: 'smooth', block: 'start' });
        await loadProductPhotos(id);
        document.getElementById('product-photos-block').style.display = 'block';
    } catch (e) { toast(e.message, true); }
}

async function loadProductPhotos(productId) {
    try {
        const photos = await api('/photo?product_id=' + productId);
        const root = document.getElementById('product-photos');
        if (!photos.length) { root.innerHTML = '<p class="hint">Фотографий нет</p>'; return; }
        root.innerHTML = photos.map((p, i) => `
          <div class="photo-thumb-wrap">
            <img src="/photo/${p.id}" class="photo-thumb">
            ${i === 0 ? '<span class="badge">главная</span>' : ''}
            <button onclick="deletePhoto('${p.id}', '${productId}')">×</button>
          </div>
        `).join('');
    } catch (e) { toast(e.message, true); }
}

async function deletePhoto(photoId, productId) {
    if (!confirm('Удалить фото?')) return;
    try {
        await api('/photo/' + photoId, { method: 'DELETE' });
        toast('Фото удалено');
        await loadProductPhotos(productId);
        loadProducts();
    } catch (e) { toast(e.message, true); }
}

document.getElementById('upload-photo').onsubmit = async (e) => {
    e.preventDefault();
    const form = document.getElementById('add-product');
    const productId = form.querySelector('[name=id]').value;
    if (!productId) { toast('Сначала создайте товар', true); return; }
    const file = document.getElementById('photo-file').files[0];
    if (!file) return;
    const fd = new FormData();
    fd.append('file', file);
    fd.append('product_id', productId);
    try {
        await api('/photo', { method: 'POST', body: fd });
        toast('Фото загружено');
        e.target.reset();
        await loadProductPhotos(productId);
        loadProducts();
    } catch (err) { toast(err.message, true); }
};

async function deleteProduct(id) {
    if (!confirm('Удалить товар?')) return;
    try {
        await api('/product/' + id, { method: 'DELETE' });
        toast('Удалён');
        loadProducts();
    } catch (e) { toast(e.message, true); }
}

// ---------- import ----------
document.getElementById('import-products').onsubmit = async (e) => {
    e.preventDefault();
    const fileInput = document.getElementById('import-file');
    if (!fileInput.files.length) return;
    try {
        const text = await fileInput.files[0].text();
        const products = JSON.parse(text);
        if (!Array.isArray(products)) { toast('Файл должен содержать JSON-массив', true); return; }
        const r = await api('/product/import', { method: 'POST', body: JSON.stringify(products) });
        toast(`Импорт: новых ${r.imported}, пополнено ${r.restocked}`);
        e.target.reset();
        loadProducts();
    } catch (err) { toast('Ошибка импорта: ' + err.message, true); }
};

// ---------- users modal ----------
const ROLE_NAMES = {
    '00000000-0000-0000-0000-000000000002': 'Клиент',
    '00000000-0000-0000-0000-000000000003': 'Менеджер',
    '00000000-0000-0000-0000-000000000004': 'Администратор',
};
const ALL_ROLES = Object.entries(ROLE_NAMES);

function openUsersModal() {
    if (typeof isAdmin === 'function' && !isAdmin()) {
        toast('Доступно только администратору', true);
        return;
    }
    const el = document.getElementById('users-modal');
    if (!el) return;
    el.style.display = 'flex';
    if (typeof loadUsersList === 'function') loadUsersList();
}

function closeUsersModal() {
    const el = document.getElementById('users-modal');
    if (el) el.style.display = 'none';
}

// кнопка открытия (если ещё не привязана)
const openUsersBtn = document.getElementById('open-users-btn');
if (openUsersBtn) openUsersBtn.onclick = openUsersModal;
document.getElementById('open-users-btn').onclick = openUsersModal;

async function loadUsersList() {
    try {
        const users = await api('/user?limit=200');
        const root = document.getElementById('users-list');
        if (!users || !users.length) { root.innerHTML = '<p class="hint">Пользователей нет</p>'; return; }
        root.innerHTML = users.map(u => `
          <div class="row">
            <span class="grow">
              ${escapeHtml(u.name)} · ${escapeHtml(u.email)}
              ${u.phone ? `<div class="meta" style="font-size:.85rem">📞 ${escapeHtml(u.phone)}</div>` : ''}
              <div class="meta" style="font-size:.75rem">Баланс: ${(u.balance ?? 0).toFixed(2)} ₽</div>
            </span>
            <select onchange="setUserRole('${u.id}', this.value)">
              ${ALL_ROLES.map(([id, name]) =>
                `<option value="${id}" ${id === u.role_id ? 'selected' : ''}>${name}</option>`).join('')}
            </select>
            <button onclick="topUpUser('${u.id}')">Пополнить</button>
            <button onclick="deleteUser('${u.id}')">Удалить</button>
          </div>`).join('');
    } catch (e) { toast(e.message, true); }
}
document.getElementById('load-users').onclick = loadUsersList;

async function setUserRole(userId, roleId) {
    try {
        await api('/user/' + userId + '/role', {
            method: 'PUT',
            body: JSON.stringify({ role_id: roleId }),
        });
        toast('Роль обновлена');
    } catch (e) { toast(e.message, true); }
}

async function topUpUser(userId) {
    const amount = parseFloat(prompt('Сумма пополнения, ₽:', '500'));
    if (!amount || amount <= 0) return;
    try {
        await api('/user/' + userId + '/balance', {
            method: 'POST',
            body: JSON.stringify({ amount }),
        });
        toast('Баланс пополнен');
        loadUsersList();
    } catch (e) { toast(e.message, true); }
}

async function deleteUser(userId) {
    if (!confirm('Удалить пользователя?')) return;
    try {
        await api('/user/' + userId, { method: 'DELETE' });
        toast('Пользователь удалён');
        loadUsersList();
    } catch (e) { toast(e.message, true); }
}

document.getElementById('create-user-form').onsubmit = async (e) => {
    e.preventDefault();
    const data = Object.fromEntries(new FormData(e.target));
    try {
        await api('/user', { method: 'POST', body: JSON.stringify(data) });
        toast('Пользователь создан');
        e.target.reset();
        loadUsersList();
    } catch (err) { toast(err.message, true); }
};

// ---------- start ----------
loadProducts();
updateHeaderAuth();
updateAdminTab();
updateTabs();
renderBalance();