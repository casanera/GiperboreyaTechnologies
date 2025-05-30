const API_BASE_URL = '/api/v1/users';

const userForm = document.getElementById('userForm');
const userIdInput = document.getElementById('userId');
const nameInput = document.getElementById('name');
const emailInput = document.getElementById('email');
const usersTableBody = document.getElementById('usersTableBody');
const clearFormButton = document.getElementById('clearFormButton');

let isEditing = false;

async function fetchUsers() {
    try {
        const response = await fetch(API_BASE_URL);
        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status} ${response.statusText}`);
        }
        const users = await response.json();
        displayUsers(users || []);
    } catch (error) {
        console.error('Ошибка при загрузке пользователей:', error);
        usersTableBody.innerHTML = `<tr><td colspan="4" style="color:red; text-align:center;">Не удалось загрузить пользователей: ${error.message}</td></tr>`;
    }
}

async function createUser(user) {
    try {
        const response = await fetch(API_BASE_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(user),
        });
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: response.statusText }));
            throw new Error(`Ошибка HTTP ${response.status}: ${errorData.message || response.statusText}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Ошибка при создании пользователя:', error);
        alert(`Не удалось создать пользователя: ${error.message}`);
        return null;
    }
}

async function updateUser(id, user) {
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(user),
        });
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: response.statusText }));
            throw new Error(`Ошибка HTTP ${response.status}: ${errorData.message || response.statusText}`);
        }
        return await response.json();
    } catch (error) {
        console.error(`Ошибка при обновлении пользователя ${id}:`, error);
        alert(`Не удалось обновить пользователя: ${error.message}`);
        return null;
    }
}

async function deleteUser(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`, {
            method: 'DELETE',
        });
        if (response.status === 204) {
            return true; 
        }
        if (!response.ok) {
            const errorText = await response.text();
            let errorMessage = response.statusText;
            if (errorText) {
                try {
                    const errorData = JSON.parse(errorText);
                    errorMessage = errorData.message || errorText;
                } catch (e) {
                    errorMessage = errorText;
                }
            }
            throw new Error(`Ошибка HTTP ${response.status}: ${errorMessage}`);
        }
        return true; 
    } catch (error) {
        console.error(`Ошибка при удалении пользователя ${id}:`, error);
        alert(`Не удалось удалить пользователя: ${error.message}`);
        return false;
    }
}

function displayUsers(users) {
    usersTableBody.innerHTML = '';
    if (!users || users.length === 0) {
        usersTableBody.innerHTML = '<tr><td colspan="4">Пользователи не найдены.</td></tr>';
        return;
    }
    users.forEach(user => {
        const row = usersTableBody.insertRow();
        row.innerHTML = `
            <td>${user.id}</td>
            <td>${escapeHTML(user.name)}</td>
            <td>${escapeHTML(user.email)}</td>
            <td class="actions">
                <button class="edit-btn" data-id="${user.id}" data-name="${escapeHTML(user.name)}" data-email="${escapeHTML(user.email)}">Редактировать</button>
                <button class="delete-btn" data-id="${user.id}">Удалить</button>
            </td>
        `;
    });
}

function escapeHTML(str) {
    if (typeof str !== 'string') return '';
    return str.replace(/[&<>"']/g, function (match) {
        const escape = {
            '&': '&',
            '<': '<',
            '>': '>',
            '"': '"',
            "'": '''
        };
        return escape[match];
    });
}


userForm.addEventListener('submit', async (event) => {
    event.preventDefault();
    const name = nameInput.value.trim();
    const email = emailInput.value.trim();
    const id = userIdInput.value;

    if (!name || !email) {
        alert('Имя и Email обязательны для заполнения.');
        return;
    }

    const userData = { name, email };
    let result;
    if (isEditing && id) {
        result = await updateUser(id, userData);
    } else {
        result = await createUser(userData);
    }
    if (result) {
        resetForm();
        fetchUsers();
    }
});

usersTableBody.addEventListener('click', async (event) => {
    const target = event.target;
    if (target.classList.contains('edit-btn')) {
        userIdInput.value = target.dataset.id;
        nameInput.value = target.dataset.name;
        emailInput.value = target.dataset.email;
        isEditing = true;
        clearFormButton.style.display = 'inline-block';
        userForm.querySelector('button[type="submit"]').textContent = 'Обновить';
        nameInput.focus();
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }
    if (target.classList.contains('delete-btn')) {
        const id = target.dataset.id;
        if (confirm(`Вы уверены, что хотите удалить пользователя с ID ${id}?`)) {
            const success = await deleteUser(id);
            if (success) {
                fetchUsers();
            }
        }
    }
});

clearFormButton.addEventListener('click', () => {
    resetForm();
});

function resetForm() {
    userForm.reset();
    userIdInput.value = '';
    isEditing = false;
    clearFormButton.style.display = 'none';
    userForm.querySelector('button[type="submit"]').textContent = 'Сохранить';
}

document.addEventListener('DOMContentLoaded', () => {
    fetchUsers();
});