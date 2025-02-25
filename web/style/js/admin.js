const userContainer = document.querySelector('.users__part__inner');

function formUserItem(id, name, space) {
    const userBlock = document.createElement('div');
    userBlock.className = 'user__element';
    userBlock.innerHTML = `
        <div class="meta">
            <img src="/profile?get=picture&id=${id}" alt="" class="user__icon">
            <h2>${name}</h2>
            <h2>Storage: ${((space/1024)/1024).toFixed(2)} Gb</h2>
        </div>
        <div class="actions">
            <img src="style/icons/trash.svg" alt="" class="actions__icon" onclick="deleteUserWarn(${id})">
            <img src="style/icons/pen.png" alt="" class="actions__icon" onclick="showFields(${id})">
        </div>
    `;

    userContainer.appendChild(userBlock);
}

function deleteUserWarn(id) {
    const warn = document.querySelector('.delete-warn');
    warn.classList.add('active');
    const button = document.querySelector('.delete-warn .fields .fields__inner .fields__btn');
    button.setAttribute('onclick', `deleteFile(${id})`);
}

function deleteUser(id) {
    fetch('/users?id=' + id, {
        method: 'DELETE',
    }).then(response => {
        if (response.ok) {
            location.reload(true);
        } else {
            console.log(response.text());
        }
    }).catch(error => {
        console.log(error);
    });
}

function showFields(id) {
    if (id === -1) document.querySelector('.fields__frame').classList.add('active', 'newUser');
    else {
        const frame = document.querySelector('.fields__frame');
        frame.classList.add('active', 'changeUser');
        frame.id = id;
    }
}

function sendFields(e) {
    e.preventDefault();
    const form = document.querySelector('.fields__form');
    const fieldsFrame = document.querySelector('.fields__frame');
    const inputs = form.querySelectorAll('.fields__input');

    if (fieldsFrame.classList.contains('newUser')) {
        const formData = new FormData(form);
        let flag = true;
        inputs.forEach(input => {
            if (input.value.trim() === '') {
                input.classList.add("input-red");
                flag = false;
            }
        });
        if (!flag) return;

        fetch('/users', {
            method: 'POST',
            body: formData,
        }).then(response => {
            if (response.ok) {
                location.reload(true);
            } else {
                inputs.forEach(input => {
                    input.classList.add("input-red");
                });
                console.log(response.text());
            }
        }).catch(error => {
            console.log(error);
        });
    }

    if (fieldsFrame.classList.contains('changeUser')) {
        const formData = new FormData(form);
        formData.append('id', fieldsFrame.id);

        fetch('/users', {
            method: 'PUT',
            body: formData,
        }).then(response => {
            if (response.ok) {
                location.reload(true);
            } else {
                console.log(response.text());
            }
        }).catch(error => {
            console.log(error);
        });
    }
}

fetch('/users?page=1&size=20', {
    method: 'GET',
}).then(response => {
    if (response.ok) {
        return response.json();
    } else {
        console.log(response.text());
    }
}).then(users => {
    users.forEach(user => {
        formUserItem(user.Id, user.Username, user.Space);
    });
}).catch(error => {
    console.log(error);
});