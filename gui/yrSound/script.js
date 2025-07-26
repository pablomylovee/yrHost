// authentication
const go_on = (addAuth) => {
	if (addAuth) {
		username = usernameInput.value;
		password = passwordInput.value;
		use_auth = true;
	} else use_auth = false;
	window.location.href = `./options/songs?username=${username}&password=${password}`;
};

let use_auth;
let username, password;
const authenticationDiv = document.getElementById("authentication");
const usernameInput = document.getElementById('username-input');
const passwordInput = document.getElementById('password-input');
const authenticateButton = document.getElementById('authenticateButton');
const dim_frame = document.getElementById("dim-frame");

fetch("/users-qm")
.then(response => {
	if (!response.ok) go_on(false);
	else authenticationDiv.style.display = "flex";
});

authenticateButton.addEventListener("click", () => {
	fetch(`/isUser-qm?username=${encodeURIComponent(usernameInput.value)}&password=${encodeURIComponent(passwordInput.value)}`)
	.then(response => {
		if (response.ok) go_on(true);
		else {
			document.querySelector('#authentication > span').textContent = "Invalid login.";
			setTimeout(() => {
				document.querySelector('#authentication > span').textContent = "Login to see your music.";
			}, 1000);
		}
	})
});