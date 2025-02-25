let userGlobal;
let fileHighlightBuffer;

function focusSearch() {
	document.querySelector('.search__input').focus();
}

function burgerActive() {
	document.querySelector('.burger__frame').classList.add('active');
}

function hideBurger(event) {
	if (event.target === event.currentTarget) {
		document.querySelector('.burger__frame').classList.remove('active');
	}
}

const menu = document.querySelector('.add__context');
const addButton = document.querySelector('.add__btn__container');
function showAddContext(event) {
	menu.style.bottom = `75px`;
	menu.style.right = `75px`;
	menu.classList.add('active');
}
function hideAddContext() {
	menu.classList.remove('active');
	document.querySelector('.file__context').classList.remove('active');
	if (fileHighlightBuffer) fileHighlightBuffer.classList.remove('highlighted');
}

const context = document.querySelector('.file__context');
document.addEventListener("click", e => {
	hideAddContext();
});
menu.addEventListener("click", event => {
	event.stopPropagation();
});
addButton.addEventListener("click", event => {
	event.stopPropagation();
});
context.addEventListener("click", event => {
	event.stopPropagation();
});

function showContextFile(event, id) {
	if (fileHighlightBuffer) fileHighlightBuffer.classList.remove('highlighted');
	fileHighlightBuffer = event.target;
	event.preventDefault();
	event.target.classList.add('highlighted');

	const screenWidth = window.innerWidth;
	const screenHeight = window.innerHeight;
	const menuWidth = context.offsetWidth;
	const menuHeight = context.offsetHeight;

	let x = event.pageX;
	let y = event.pageY;

	if (x + menuWidth > screenWidth) {
		x = screenWidth - menuWidth - 10;
	}
	if (y + menuHeight > screenHeight) {
		y = screenHeight - menuHeight - 10;
	}
	context.classList.add('active');
	context.style.left = `${x}px`;
	context.style.top = `${y}px`;

	document.getElementById('download__file').setAttribute('onclick', `hideAddContext(); downloadFile("${id}")`);
	document.getElementById('delete__file').setAttribute('onclick', `hideAddContext(); deleteFileWarn("${id}")`);
	document.getElementById('rename__file').setAttribute('onclick', `hideAddContext(); renameFileField("${id}")`);
	document.getElementById('fav__file').setAttribute('onclick', `hideAddContext(); favFile("${id}", true)`);
	if (event.target.classList.contains('publicFile')) {
		document.getElementById('public__file').setAttribute('onclick', `hideAddContext(); publicFile("${id}", false)`);
		document.getElementById('public__file').textContent = "Make private";
	} else {
		document.getElementById('public__file').setAttribute('onclick', `hideAddContext(); publicFile("${id}", true)`);
		document.getElementById('public__file').textContent = "Make public";
	}
}

function hideFields(event) {
	if (event.target === event.currentTarget) {
		event.target.classList.remove('active');
		event.target.classList.remove('newUser');
		event.target.classList.remove('changeUser');
	}
}

function downloadFile(id) {
	location.href = '/files?id='+id;
}

function renameFileField(id) {
	const field = document.querySelector('.rename-field');
	field.classList.add('active');
	const form = field.querySelector('.fields__form');
	form.setAttribute('onsubmit', `renameFile(event, "${id}")`);

	form.querySelector('.fields__input').value = document.querySelector(`h2[class='${id}']`).textContent;
}

function renameFile(event, id) {
	event.preventDefault();
	const field = document.querySelector('.rename-field');
	field.classList.remove('active');
	const form = field.querySelector('.fields__form');
	const formData = new FormData(form);
	formData.append('id', id);

	fetch('/files', {
		method: 'PUT',
		body: formData,
	}).then(response => {
		if (response.ok) {
			document.getElementById('files-private-container').innerHTML = ``;
			document.getElementById('files-public-container').innerHTML = ``;
			updateProfile();
			showAllFiles();
		} else {
			console.log(response.text());
		}
	}).catch(error => {
		console.log(error);
	});
}

function publicFile(id, flag) {
	const formData = new FormData();
	formData.append("id", id);
	if (flag) formData.append("public", "true");
	if (!flag) formData.append("public", "false");

	fetch('/files', {
		method: 'PUT',
		body: formData,
	}).then(response => {
		if (response.ok) {
			document.getElementById('files-private-container').innerHTML = ``;
			document.getElementById('files-public-container').innerHTML = ``;
			updateProfile();
			showAllFiles();
		} else {
			console.log(response.text());
		}
	}).catch(error => {
		console.log(error);
	});
}

function uploadFile() {
	const fileInput = document.getElementById("file-input");
	fileInput.click();

	fileInput.onchange = async () => {
		const files = fileInput.files;
		if (!files) return;

		const formData = new FormData();
		for (let i=0; i < files.length; i++) {
			formData.append("files", files[i]);
		}

		await fetch("/files", {
			method: "POST",
			body: formData,
		}).then(response => {
			if (response.ok) {
				document.getElementById('files-private-container').innerHTML = ``;
				document.getElementById('files-public-container').innerHTML = ``;
				updateProfile();
				showAllFiles();
			} else if (response.status === 403) {
				document.querySelector('.space-warn').classList.add('active');
			} else {
				console.log(response.text());
			}
		}).catch(error => {
			console.error(error);
		});
	};
}

function formFileElement(file, privateFlag) {
	let fileContainer;

	const fileBlock = document.createElement('div');
	fileBlock.className = 'file__element';

	if (privateFlag) fileContainer = document.getElementById('files-private-container');
	if (!privateFlag) {
		fileContainer = document.getElementById('files-public-container');
		fileBlock.classList.add('publicFile');
	}

	const date = new Date(file["Update"]);
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, "0");
	const day = String(date.getDate()).padStart(2, "0");
	const hours = String(date.getHours()).padStart(2, "0");
	const minutes = String(date.getMinutes()).padStart(2, "0");

	const formattedDate = `${year}.${month}.${day} ${hours}:${minutes}`;
	const extension = file["Name"].slice(file["Name"].lastIndexOf('.') + 1).toLowerCase();

	let size;
	if (file["Size"] > 1024 * 1024) {
		file["Size"] = ((file["Size"]/1024) / 1024).toFixed(2);
		size = "Gb";
	}
	else if (file["Size"] > 1024) {
		file["Size"] = (file["Size"]/1024).toFixed(2);
		size = "Mb";
	} else {
		file["Size"] = file["Size"].toFixed(0);;
		size = "Kb";
	}

	if (privateFlag || file["Owner"] === userGlobal["id"]) {
		fileBlock.setAttribute('oncontextmenu', `showContextFile(event, \"${file["Id"]}\")`);
		fileBlock.innerHTML = `
        <div class="meta">
			<img src="style/icons/file_icons/${extension}.png" alt="" class="file__icon">
			<h2 class="${file["Id"]}">${file["Name"]}</h2>
		</div>
		<div class="actions">
			<h2>${file["Size"]} ${size}</h2>
			<h2>Uploaded: ${formattedDate}</h2>
			<img src="style/icons/trash.svg" alt="" class="actions__icon" onclick="deleteFileWarn('${file["Id"]}')">
			<a href="/files?id=${file["Id"]}"><img src="style/icons/download.svg" alt="" class="actions__icon"></a>
		</div>
    `;
	} else {
		fileBlock.innerHTML = `
        <div class="meta">
			<img src="style/icons/file_icons/${extension}.png" alt="" class="file__icon">
			<h2>${file["Name"]}</h2>
		</div>
		<div class="actions">
			<h2>${file["Size"]} ${size}</h2>
			<h2>Uploaded: ${formattedDate}</h2>
			<a href="/files?id=${file["Id"]}"><img src="style/icons/download.svg" alt="" class="actions__icon"></a>
		</div>
    `;
	}
	if (file["Favourite"] === true) fileBlock.classList.add('favourite');

	fileContainer.appendChild(fileBlock);
}

function deleteFileWarn(id) {
	const warn = document.querySelector('.delete-warn');
	warn.classList.add('active');
	const button = document.querySelector('.delete-warn .fields .fields__inner .fields__btn');
	button.setAttribute('onclick', `deleteFile('${id}')`);
}

function deleteFile(id) {
	document.querySelector('.delete-warn').classList.remove('active');

	fetch('/files?id=' + id, {
		method: 'DELETE',
	}).then(response => {
		if (response.ok) {
			document.getElementById('files-private-container').innerHTML = ``;
			document.getElementById('files-public-container').innerHTML = ``;
			updateProfile();
			showAllFiles();
		} else {
			console.log(response.text());
		}
	}).catch(error => {
		console.log(error);
	});
}

const usernameElement = document.querySelector('.profile h1');
const spaceElement = document.querySelector('.burger__menu__inner h2');

function updateProfile() {
	fetch('/profile', {
		method: 'GET',
	}).then(response => {
		if (response.ok) {
			return response.json();
		} else {
			console.log(response.text());
		}
	}).then(data => {
		userGlobal = data;
		usernameElement.textContent = data["name"];
		spaceElement.textContent = `${((data["space_occupied"] / 1024) / 1024).toFixed(2)} / ${((data["space_available"] / 1024) / 1024).toFixed(2)} Gb Occupied`;
	}).catch(error => {
		console.error(error);
	});
}

function showAllFiles() {
	fetch('/files?page=1&size=20', {
		method: 'GET',
	}).then(response => {
		if (response.ok) {
			return response.json();
		} else {
			console.log(response.text());
		}
	}).then(files => {
		files.forEach(file => {
			if (file["Public"] === false) {
				formFileElement(file, true);
			} else {
				formFileElement(file, false);
			}
		});

		const privateBlock = document.getElementById('files-private-container');
		if (privateBlock.childElementCount === 0) {
			const h1 = document.createElement('h1');
			h1.id = 'zero-upload-files';
			h1.textContent = 'OH! YOU DIDN\'T UPLOAD ANY FILES YET!';
			h1.style.margin = '0';
			privateBlock.appendChild(h1);
		} else {
			const upLabel = document.getElementById('zero-upload-files');
			if (upLabel) upLabel.remove();
		}

		const publicBlock = document.getElementById('files-public-container');
		const label = document.getElementById('files-public-label');
		if (publicBlock.childElementCount === 0) {
			label.style.display = 'none';
		} else {
			label.style.display = 'block';
		}
	}).catch(error => {
		console.error(error);
	});
}

function hideSpaceWarn() {
	document.querySelector('.space-warn').classList.remove('active');
}

updateProfile();
showAllFiles();