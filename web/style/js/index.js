let setIconFlag = true;
let burgerFlag = true;
const passIcon = document.querySelector('.pass-icon');
const inputPass = document.querySelector('.input-pass');
const loginInput = document.querySelectorAll('.login__input');
const loginForm = document.querySelector('.login__form');
const headerNav = document.querySelector('.header__nav');

function setIcon() {
	if (setIconFlag) {
		passIcon.setAttribute('src', 'style/icons/eye.svg');
		inputPass.setAttribute('type', 'text');
	} else {
		passIcon.setAttribute('src', 'style/icons/eye-hide.svg');
		inputPass.setAttribute('type', 'password');
	}
	setIconFlag = !setIconFlag;
}
function burgerActive() {
	if(burgerFlag){
		headerNav.classList.add("header__nav-active");
	} else {
		headerNav.classList.remove("header__nav-active");
	}
	burgerFlag = !burgerFlag;
}

function sendLogin(event) {
	event.preventDefault();
	const formData = new FormData(loginForm);

	fetch('/login', {
        method: 'POST',
        body: formData
    })
    .then(response => {
    	if (response.status === 401) {
    		loginInput[0].classList.add("input-red");
			loginInput[1].classList.add("input-red");
			passIcon.classList.add("pass-icon-red");
    	}
    	if (response.ok) {
    		window.location.href = '/dashboard';
    	}
    })
    .then(data => {
        console.log(data);
    })
    .catch(error => {
        console.log(error);
    });
}