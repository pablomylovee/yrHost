// authentication
const go_on = (addAuth) => {
	if (addAuth) {
		username = usernameInput.value;
		password = passwordInput.value;
		use_auth = true;
	} else use_auth = false;
	sessionStorage.setItem("username", username);
	sessionStorage.setItem("password", password);
	authenticationDiv.style.display = 'none';
	document.getElementById('mainContent').style.display = 'block';
	get_files();
}

export let use_auth;
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
				document.querySelector('#authentication > span').textContent = "Login to see your files.";
			}, 1000);
		}
	})
});

// container elements
const imgcont = document.getElementById("imgcont");
const videocont = document.getElementById("videocont");
const audiocont = document.getElementById("audiocont");
const textcont = document.getElementById("textcont");

dim_frame.addEventListener("click", () => {
	dim_frame.style.display = "none";

	imgcont.src = "";
	imgcont.style.display = "none";
	videocont.src = "";
	videocont.style.display = "none";
	audiocont.src = "";
	audiocont.style.display = "none";
	textcont.textContent = "Loading, please wait...";
	textcont.style.display = "none";
});
// main content
const pathspan = document.querySelector("#path > span");
const files = document.getElementById('files');
const gb = document.getElementById('gb');
const refresh = document.getElementById('refresh');

gb.addEventListener('click', () => {
	if (sessionStorage.getItem("current_dir") !== '') {
		let x = sessionStorage.getItem("current_dir").split('/');
		x.pop(); x = x.join("/");

		get_files(x);
	}
});
refresh.addEventListener('click', () => get_files(sessionStorage.getItem('current_dir')));

export const getFetchBall = (ballModel, notes) => {
	let resBall = `/${ballModel}`;
	if (use_auth) resBall = `${resBall}?username=${username}&password=${password}`;
	if (typeof notes !== 'undefined') {
		if (use_auth) resBall = `${resBall}&${notes}`;
		else resBall = `${resBall}?${notes}`;
	}

	return resBall;
}

export const get_files = (dir) => {
	let reach_to;
	if (typeof dir == "undefined" || dir == '') {
		reach_to = getFetchBall('get-files');
		sessionStorage.setItem("current_dir", '');
		pathspan.textContent = "~";
	} else {
		reach_to = getFetchBall('get-files', `dir=${encodeURIComponent(dir.toString())}`);
		sessionStorage.setItem("current_dir", dir);
		pathspan.textContent = `~/${dir}`;
	}

	fetch(reach_to)
	.then(response => {
		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}
		return response.json();
	})
	.then(data => {
		for (const entry of Array.from(files.childNodes)) entry.remove();

		if (data == null) { return; }
		for (const entry of data) {
			const entry_div = document.createElement("div");
			entry_div.classList.add("fileEntry");
			const entry_span = document.createElement('span');
			entry_span.textContent = `${entry.name}`;
			if (entry.type == "d") {
				entry_span.style.color = '#4499ff';
				entry_span.style.fontWeight = "bolder";
			}
			const delete_button = document.createElement('button');
			delete_button.classList.add("entry-deletebutton");
			const delete_img = document.createElement("img");
			delete_img.src = "./vectors/trash.svg";
			delete_img.width = "25"; delete_img.height = "25";
			delete_button.appendChild(delete_img);
			const rename_button = document.createElement('button');
			const rename_img = document.createElement("img");
			rename_img.src = "./vectors/rename.svg";
			rename_img.width = "25"; rename_img.height = "25";
			rename_button.appendChild(rename_img);
			rename_button.classList.add("entry-renamebutton");
			const download_button = document.createElement('button');
			const download_img = document.createElement("img");
			download_img.src = "./vectors/download.svg";
			download_img.width = "25"; download_img.height = "25";
			download_button.appendChild(download_img);
			download_button.classList.add("entry-downloadbutton");

			entry_span.addEventListener('click', () => {
				console.log('clicked');

				if (sessionStorage.getItem("current_dir") == '' && entry.type == "d") {
					get_files(entry.name);
				} else if (entry.type == "d") {
					get_files(sessionStorage.getItem("current_dir")+"/"+entry.name);
				} else if (entry.type == "f" && sessionStorage.getItem("current_dir") == '') {
					fetch(getFetchBall('get-content', `path=${entry.name}`))
					.then(response => {
						const contentType = response.headers.get('Content-Type');
					
						return response.blob().then(blob => {
							return {blob, contentType}
						});
					})
					.then(async({blob, contentType}) => {
						dim_frame.style.animation = "fade 300ms forwards reverse";
						dim_frame.style.display = "block";
						if (contentType.startsWith("image/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";

							const url = URL.createObjectURL(blob);
							imgcont.src = url;

							const returnedimg = new Image();
							returnedimg.src = url;
							await returnedimg.decode();
							imgcont.style.aspectRatio = `${returnedimg.naturalWidth} / ${returnedimg.naturalHeight}`;
							if (returnedimg.naturalWidth / returnedimg.naturalHeight <= 1.05) {
								imgcont.style.height = "80%";
								imgcont.style.width = "";
							} else {
								imgcont.style.width = "80%";
								imgcont.style.height = "";
							}

							imgcont.style.display = "block";
						} else if (contentType.startsWith("video/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";
						
							const url = URL.createObjectURL(blob);
							videocont.src = url;
							const loadmetadata = () => {
								videocont.style.aspectRatio = `${videocont.videoWidth} / ${videocont.videoHeight}`;
								if (videocont.videoWidth / videocont.videoHeight <= 1.05) {
									videocont.style.height = "80%";
									videocont.style.width = "";
								} else {
									videocont.style.width = "80%";
									videocont.style.height = "";
								}

								dim_frame.style.display = "block";
								videocont.style.display = "block";
								videocont.removeEventListener("loadedmetadata", loadmetadata);	
							}
							videocont.addEventListener("loadedmetadata", loadmetadata);
						} else if (contentType.startsWith("audio/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";

							const url = URL.createObjectURL(blob);
							audiocont.src = url;

							dim_frame.style.display = "block";
							audiocont.style.display = "block";
						} else {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";
							
							blob.text().then(text => { textcont.textContent = text; });

							dim_frame.style.display = "block";
							textcont.style.display = "block";
						}
					});
				} else if (entry.type == "f") {
					const dir = sessionStorage.getItem("current_dir");
					const input_path = dir ? `${encodeURIComponent(dir)}~%2F~${encodeURIComponent(entry.name)}` : encodeURIComponent(entry.name);
					fetch(getFetchBall('get-content', `path=${input_path}`))
					.then(response => {
						const contentType = response.headers.get('Content-Type');
					
						return response.blob().then(blob => {
							return {blob, contentType}
						});
					})
					.then(async({blob, contentType}) => {
						dim_frame.style.animation = "fade 150ms forwards reverse";
						dim_frame.style.display = "block";
						if (contentType.startsWith("image/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";

							const url = URL.createObjectURL(blob);
							imgcont.src = url;

							const returnedimg = new Image();
							returnedimg.src = url;
							await returnedimg.decode();
							imgcont.style.aspectRatio = `${returnedimg.naturalWidth} / ${returnedimg.naturalHeight}`;
							if (returnedimg.naturalWidth / returnedimg.naturalHeight <= 1.05) {
								imgcont.style.height = "80%";
								imgcont.style.width = "";
							} else {
								imgcont.style.width = "80%";
								imgcont.style.height = "";
							}

							imgcont.style.display = "block";
						} else if (contentType.startsWith("video/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";
						
							const url = URL.createObjectURL(blob);
							videocont.src = url;
							const loadmetadata = () => {
								videocont.style.aspectRatio = `${videocont.videoWidth} / ${videocont.videoHeight}`;
								if (videocont.videoWidth / videocont.videoHeight <= 1.05) videocont.style.height = "80%";
								else videocont.style.width = "80%";

								dim_frame.style.display = "block";
								videocont.style.display = "block";
								videocont.removeEventListener("loadedmetadata", loadmetadata);	
							}
							videocont.addEventListener("loadedmetadata", loadmetadata);
						} else if (contentType.startsWith("audio/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";

							const url = URL.createObjectURL(blob);
							audiocont.src = url;

							dim_frame.style.display = "block";
							audiocont.style.display = "block";
						} else {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";
							
							blob.text().then(text => { textcont.textContent = text; });

							dim_frame.style.display = "block";
							textcont.style.display = "block";
						}
					});
				}	
			});
			delete_button.addEventListener("mouseenter", () => { delete_img.src = "./vectors/trash-hover.svg"; });
			delete_button.addEventListener("mouseleave", () => { delete_img.src = "./vectors/trash.svg"; });
			delete_button.addEventListener('click', async() => {
				if (window.confirm(`Are you sure you want to delete '${entry.name}'?`)) {
					if (sessionStorage.getItem("current_dir") == '') {
						fetch(getFetchBall('delete-file', `name=${entry.name}`))
						.then(response => {
							if (response.status === 404) {
								window.alert(`No such file as ${entry.name}!`);
							}
							if (response.status === 302) get_files(sessionStorage.getItem('current_dir'));
						})
					} else {
						const path = sessionStorage.getItem("current_dir").replaceAll("/", "~%2F~");
						fetch(getFetchBall('delete-file', `name=${path}~%2F~${entry.name}`))
						.then(response => {
							if (response.status === 404) {
								window.alert(`No such file as ${entry.name}!`);
							}
							if (response.status === 302) get_files(sessionStorage.getItem('current_dir'));
						})
					}
				}
			});

			rename_button.addEventListener("click", () => {
				const new_name = window.prompt(`Enter a new name for '${entry.name}'.`, entry.name);
				if (new_name === null || new_name.trim() === "") return;

				const currentDir = sessionStorage.getItem("current_dir");

				if (currentDir === null || currentDir.trim() === '') {
					fetch(getFetchBall('rename-file', `path=${entry.name}&newname=${new_name}`))
						.then(response => {
							if (response.ok) get_files(currentDir);
						});
				} else {
					const encodedPath = currentDir.replaceAll("/", "~%2F~");
					fetch(getFetchBall('rename-file', `path=${encodedPath}~%2F~${entry.name}&newname=${new_name}`))
						.then(response => {
							if (response.ok) get_files(currentDir);
						});
				}
			});
			download_button.addEventListener("click", () => {
				const input_path = sessionStorage.getItem("current_dir")
					? `${encodeURIComponent(sessionStorage.getItem("current_dir").split("/").join("~/~"))}~%2F~${encodeURIComponent(entry.name)}`
					: encodeURIComponent(entry.name);
				fetch(getFetchBall("get-content", `path=${input_path}`))
				.then(response => {
					if (response.ok) return response.blob();
				})
				.then(blob => {
					const url = URL.createObjectURL(blob);
					const link = document.createElement("a");
					link.style.display = "none";
					link.href = url;
					link.target = "_blank";
					link.download = entry.name;
					link.click();
					link.remove();
				});
			})

			entry_div.appendChild(entry_span);
			entry_div.appendChild(delete_button);
			entry_div.appendChild(rename_button);
			entry_div.appendChild(download_button);
			files.appendChild(entry_div);
		}
	})
	.catch(error => {
		console.error('There was a problem with the fetch operation:', error);
	});
}
