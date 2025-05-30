

// URL нашего API
const API_BASE_URL = '/api/v1/users/';

// Получаем ссылки на элементы DOM
const userForm = document.getElementById('userForm');
const userIdInput = document.getElementById('userId');
const nameInput = document.getElementById('name');
const emailInput = document.getElementById('email');
const usersTableBody = document.getElementById('usersTableBody');
const clearFormButton = document.getElementById('clearFormButton');

let isEditing = false; // Флаг, находимся ли мы в режиме редактирования

console.log("DEBUG_SCRIPT: Скрипт script.js загружен. Переменные DOM:", 
    { userForm, userIdInput, nameInput, emailInput, usersTableBody, clearFormButton }
);

// --- ФУНКЦИИ ДЛЯ ВЗАИМОДЕЙСТВИЯ С API ---

// Функция для получения всех пользователей
async function fetchUsers() {
    console.log("DEBUG_API: fetchUsers - Начало вызова");
    try {
        const response = await fetch(API_BASE_URL);
        console.log("DEBUG_API: fetchUsers - Ответ от fetch:", response);
        if (!response.ok) {
            const errorText = await response.text(); // Попробуем получить текст ошибки
            console.error(`DEBUG_API: fetchUsers - Ошибка HTTP: ${response.status} ${response.statusText}. Тело ошибки: ${errorText}`);
            throw new Error(`Ошибка HTTP: ${response.status} ${response.statusText}. Сервер ответил: ${errorText}`);
        }
        const users = await response.json();
        console.log("DEBUG_API: fetchUsers - Получены пользователи:", users);
        displayUsers(users || []);
    } catch (error) {
        console.error('КРИТИЧЕСКАЯ ОШИБКА при загрузке пользователей (fetchUsers):', error);
        usersTableBody.innerHTML = `<tr><td colspan="4" style="color:red; text-align:center;">Не удалось загрузить пользователей: ${error.message}</td></tr>`;
    }
}

// Функция для создания пользователя
async function createUser(user) {
    console.log("DEBUG_API: createUser - Начало вызова. Данные:", user);
    try {
        const response = await fetch(API_BASE_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(user),
        });
        console.log("DEBUG_API: createUser - Ответ от fetch:", response);
        if (!response.ok) {
            const errorData = await response.json().catch(async () => ({ message: await response.text() || response.statusText }));
            console.error(`DEBUG_API: createUser - Ошибка HTTP ${response.status}:`, errorData);
            throw new Error(`Ошибка HTTP ${response.status}: ${errorData.message || response.statusText}`);
        }
        const createdUser = await response.json();
        console.log("DEBUG_API: createUser - Пользователь создан:", createdUser);
        return createdUser;
    } catch (error) {
        console.error('КРИТИЧЕСКАЯ ОШИБКА при создании пользователя (createUser):', error);
        alert(`Не удалось создать пользователя: ${error.message}`);
        return null;
    }
}

// Функция для обновления пользователя
async function updateUser(id, user) {
    console.log(`DEBUG_API: updateUser - Начало вызова. ID: ${id}, Данные:`, user);
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(user),
        });
        console.log("DEBUG_API: updateUser - Ответ от fetch:", response);
        if (!response.ok) {
            const errorData = await response.json().catch(async () => ({ message: await response.text() || response.statusText }));
            console.error(`DEBUG_API: updateUser - Ошибка HTTP ${response.status}:`, errorData);
            throw new Error(`Ошибка HTTP ${response.status}: ${errorData.message || response.statusText}`);
        }
        const updatedUser = await response.json();
        console.log("DEBUG_API: updateUser - Пользователь обновлен:", updatedUser);
        return updatedUser;
    } catch (error) {
        console.error(`КРИТИЧЕСКАЯ ОШИБКА при обновлении пользователя ${id} (updateUser):`, error);
        alert(`Не удалось обновить пользователя: ${error.message}`);
        return null;
    }
}

// Функция для удаления пользователя
async function deleteUser(id) {
    console.log(`DEBUG_API: deleteUser - Начало вызова. ID: ${id}`);
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`, {
            method: 'DELETE',
        });
        console.log("DEBUG_API: deleteUser - Ответ от fetch:", response);
        if (response.status === 204) {
            console.log(`DEBUG_API: deleteUser - Пользователь ID ${id} успешно удален (статус 204).`);
            return true; 
        }
        if (!response.ok) {
            const errorText = await response.text();
            let errorMessage = response.statusText;
            if (errorText) {
                try {
                    const errorData = JSON.parse(errorText); // Попытка распарсить как JSON
                    errorMessage = errorData.error || errorData.message || errorText; // Ищем поле error или message
                } catch (e) {
                    errorMessage = errorText; // Если не JSON, используем как текст
                }
            }
            console.error(`DEBUG_API: deleteUser - Ошибка HTTP ${response.status}: ${errorMessage}`);
            throw new Error(`Ошибка HTTP ${response.status}: ${errorMessage}`);
        }
        console.log(`DEBUG_API: deleteUser - Пользователь ID ${id} успешно удален (статус ${response.status}).`);
        return true; 
    } catch (error) {
        console.error(`КРИТИЧЕСКАЯ ОШИБКА при удалении пользователя ${id} (deleteUser):`, error);
        alert(`Не удалось удалить пользователя: ${error.message}`);
        return false;
    }
}

// --- ФУНКЦИИ ДЛЯ ОТОБРАЖЕНИЯ ДАННЫХ ---

// Функция для отображения пользователей в таблице
function displayUsers(users) {
    console.log("DEBUG_DOM: displayUsers - Начало. Получено пользователей:", users);
    usersTableBody.innerHTML = ''; // Очищаем таблицу перед обновлением

    if (!users || users.length === 0) {
        console.log("DEBUG_DOM: displayUsers - Пользователи не найдены или массив пуст.");
        usersTableBody.innerHTML = '<tr><td colspan="4">Пользователи не найдены.</td></tr>';
        return;
    }

    users.forEach(user => {
        const row = usersTableBody.insertRow();
        // ПРЯМАЯ ВСТАВКА ДАННЫХ (без escapeHTML для отладки)
        // ВНИМАНИЕ: В ПРОДАШЕНЕ ЗДЕСЬ ДОЛЖНО БЫТЬ ЭКРАНИРОВАНИЕ ДЛЯ БЕЗОПАСНОСТИ!
        row.innerHTML = `
            <td>${user.id}</td>
            <td>${user.name}</td> 
            <td>${user.email}</td>
            <td class="actions">
                <button class="edit-btn" data-id="${user.id}" data-name="${user.name}" data-email="${user.email}">Редактировать</button>
                <button class="delete-btn" data-id="${user.id}">Удалить</button>
            </td>
        `;
    });
    console.log("DEBUG_DOM: displayUsers - Таблица обновлена.");
}

// --- ОБРАБОТЧИКИ СОБЫТИЙ ---

if (userForm) {
    userForm.addEventListener('submit', async (event) => {
        console.log("DEBUG_EVENT: Обработчик userForm 'submit' СРАБОТАЛ!");
        event.preventDefault();

        console.log("DEBUG_EVENT: submit - Шаг 1: Собираем данные из формы");
        const name = nameInput.value.trim();
        const email = emailInput.value.trim();
        const id = userIdInput.value; // Может быть пустым, если это создание
        console.log(`DEBUG_EVENT: submit - Данные из формы: id='${id}', name='${name}', email='${email}', isEditing=${isEditing}`);

        if (!name || !email) {
            alert('Имя и Email обязательны для заполнения.');
            console.log("DEBUG_EVENT: submit - Ошибка валидации: Пустые имя или email");
            return;
        }

        const userData = { name, email };
        let result;

        try {
            if (isEditing && id) {
                console.log(`DEBUG_EVENT: submit - Шаг 2: Вызываем updateUser для ID ${id}`);
                result = await updateUser(id, userData);
                console.log("DEBUG_EVENT: submit - Результат updateUser:", result);
            } else {
                console.log("DEBUG_EVENT: submit - Шаг 2: Вызываем createUser");
                result = await createUser(userData);
                console.log("DEBUG_EVENT: submit - Результат createUser:", result);
            }

            if (result) { // result должен быть объектом пользователя при успехе, или true для delete
                console.log("DEBUG_EVENT: submit - Шаг 3: Операция API успешна, сбрасываем форму и обновляем список");
                resetForm();
                await fetchUsers(); // Обновляем список пользователей
            } else {
                console.log("DEBUG_EVENT: submit - Шаг 3: Операция API НЕ успешна (result is null/false/undefined или вернулась ошибка, обработанная в API функции)");
            }
        } catch (apiError) {
            console.error("КРИТИЧЕСКАЯ ОШИБКА при вызове API из обработчика формы 'submit':", apiError);
            alert("Произошла неожиданная ошибка при отправке данных. Пожалуйста, проверьте консоль.");
        }
    });
} else {
    console.error("КРИТИЧЕСКАЯ ОШИБКА: Элемент userForm не найден на странице!");
}

if (usersTableBody) {
    usersTableBody.addEventListener('click', async (event) => {
        console.log("DEBUG_EVENT: Обработчик usersTableBody 'click' СРАБОТАЛ! Цель:", event.target);
        const target = event.target;

        try { // Обернем в try-catch на всякий случай
            if (target.classList.contains('edit-btn')) {
                console.log("DEBUG_EVENT: click - Нажата кнопка 'Редактировать'");
                const id = target.dataset.id;
                const name = target.dataset.name;
                const email = target.dataset.email;
                console.log(`DEBUG_EVENT: click - Редактирование ID: ${id}, Имя: ${name}, Email: ${email}`);

                userIdInput.value = id;
                nameInput.value = name; 
                emailInput.value = email;
                isEditing = true;
                clearFormButton.style.display = 'inline-block';
                userForm.querySelector('button[type="submit"]').textContent = 'Обновить';
                nameInput.focus();
                window.scrollTo({ top: 0, behavior: 'smooth' });
            }

            if (target.classList.contains('delete-btn')) {
                console.log("DEBUG_EVENT: click - Нажата кнопка 'Удалить'");
                const id = target.dataset.id;
                console.log(`DEBUG_EVENT: click - Удаление ID: ${id}`);

                if (confirm(`Вы уверены, что хотите удалить пользователя с ID ${id}?`)) {
                    console.log(`DEBUG_EVENT: click - Пользователь подтвердил удаление ID: ${id}. Вызываем deleteUser.`);
                    const success = await deleteUser(id);
                    console.log(`DEBUG_EVENT: click - Результат deleteUser для ID ${id}:`, success);
                    if (success) {
                        console.log(`DEBUG_EVENT: click - Удаление успешно, обновляем список пользователей.`);
                        await fetchUsers(); 
                    } else {
                        console.log(`DEBUG_EVENT: click - Удаление НЕ успешно для ID: ${id}.`);
                    }
                } else {
                    console.log(`DEBUG_EVENT: click - Пользователь отменил удаление ID: ${id}.`);
                }
            }
        } catch (handlerError) {
            console.error("КРИТИЧЕСКАЯ ОШИБКА в обработчике кликов по таблице:", handlerError);
            alert("Произошла неожиданная ошибка при обработке действия. Пожалуйста, проверьте консоль.");
        }
    });
} else {
    console.error("КРИТИЧЕСКАЯ ОШИБКА: Элемент usersTableBody не найден на странице!");
}

// Обработчик для кнопки "Отмена" (сброс формы)
if (clearFormButton) {
    clearFormButton.addEventListener('click', () => {
        console.log("DEBUG_EVENT: Нажата кнопка 'Отмена'");
        resetForm();
    });
} else {
    console.warn("ПРЕДУПРЕЖДЕНИЕ: Элемент clearFormButton не найден на странице (возможно, это нормально, если он создается динамически или не всегда нужен).");
}


// Функция для сброса формы и режима редактирования
function resetForm() {
    console.log("DEBUG_FN: resetForm - Начало");
    if(userForm) userForm.reset();
    if(userIdInput) userIdInput.value = '';
    isEditing = false;
    if(clearFormButton) clearFormButton.style.display = 'none';
    if(userForm) userForm.querySelector('button[type="submit"]').textContent = 'Сохранить';
    console.log("DEBUG_FN: resetForm - Форма сброшена");
}


// --- ИНИЦИАЛИЗАЦИЯ ---
// Загружаем пользователей при первой загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    console.log("DEBUG_INIT: DOMContentLoaded - DOM полностью загружен и разобран.");
    // Проверяем, существуют ли ключевые элементы перед вызовом fetchUsers
    if (usersTableBody && userForm && nameInput && emailInput && userIdInput) {
        console.log("DEBUG_INIT: Все ключевые DOM элементы найдены. Вызываем fetchUsers().");
        fetchUsers();
    } else {
        console.error("КРИТИЧЕСКАЯ ОШИБКА ПРИ ИНИЦИАЛИЗАЦИИ: Один или несколько ключевых DOM элементов не найдены! Не могу продолжить.");
        if (!usersTableBody) console.error("Ошибка: usersTableBody не найден.");
        if (!userForm) console.error("Ошибка: userForm не найден.");
        // ... и так далее для других элементов
    }
});