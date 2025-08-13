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
};

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
		reach_to = getFetchBall(`get-files/${encodeURI(dir)}`);
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
			const contextMenuButton = document.createElement("button");
			const cmb_img = document.createElement("img");
			cmb_img.src = "./vectors/options.svg";
			cmb_img.width = "15";
			contextMenuButton.appendChild(cmb_img);
			const contextMenu = document.createElement("div");
			contextMenu.classList.add("options");
			const delete_button = document.createElement('button');
			delete_button.textContent = `Delete`;
			const delete_img = document.createElement("img");
			delete_img.src = "./vectors/trash.svg";
			delete_img.width = "23";
			delete_button.appendChild(delete_img);
			const rename_button = document.createElement('button');
			rename_button.textContent = `Rename`;
			const rename_img = document.createElement("img");
			rename_img.src = "./vectors/rename.svg";
			rename_img.width = "23";
			rename_button.appendChild(rename_img);
			const download_button = document.createElement('button');
			download_button.textContent = `Download`;
			const download_img = document.createElement("img");
			download_img.src = "./vectors/download.svg";
			download_img.width = "23";
			download_button.appendChild(download_img);
			const edit_button = document.createElement("button");
			edit_button.textContent = "Edit with yrText";
			const edit_img = document.createElement("img");
			edit_img.src = "./vectors/edit.svg";
			edit_img.width = "23";
			edit_button.appendChild(edit_img);
			const loading_img = document.createElement("img");
			loading_img.src = "./vectors/loading.svg";
			loading_img.width = "20";
			entry_span.appendChild(loading_img);

			entry_span.addEventListener('click', () => {
				console.log('clicked');

				if (sessionStorage.getItem("current_dir") == '' && entry.type == "d") {
					get_files(entry.name);
				} else if (entry.type == "d") {
					get_files(sessionStorage.getItem("current_dir")+"/"+entry.name);
				} else if (entry.type == "f") {
					loading_img.style.display = "block";
					const url = `/yrFiles/files/${encodeURI(entry["relative-path"])}?username=${username}&password=${password}`;
					fetch(url, {method: "HEAD"})
					.then(async(response) => {
						const contentType = response.headers.get('Content-Type');
						dim_frame.style.animation = "fade 300ms forwards reverse";
						dim_frame.style.display = "block";
						if (contentType.startsWith("image/")) {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";
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
							audiocont.src = url;

							dim_frame.style.display = "block";
							audiocont.style.display = "block";
						} else {
							imgcont.style.display = "none";
							videocont.style.display = "none";
							audiocont.style.display = "none";
							textcont.style.display = "none";
							
							await fetch(url).then(response => response.text())
							.then(text => textcont.textContent = text)

							dim_frame.style.display = "block";
							textcont.style.display = "block";
						}
						loading_img.style.display = "none";
					});
				}	
			});
			delete_button.addEventListener("mouseenter", () => { delete_img.src = "./vectors/trash-hover.svg"; });
			delete_button.addEventListener("mouseleave", () => { delete_img.src = "./vectors/trash.svg"; });
			delete_button.addEventListener('click', async() => {
				if (window.confirm(`Are you sure you want to delete '${entry.name}'?`))
					fetch(getFetchBall('delete-file', `name=${encodeURIComponent(entry["relative-path"].replaceAll("/", "~/~"))}`))
					.then(response => {
						if (response.status === 404) window.alert(`No such file as ${entry.name}!`);
						if (response.status === 302) get_files(sessionStorage.getItem('current_dir'));
					});
				contextMenu.style.display = "";
				contextMenuButton.style.backgroundColor = ""
			});

			rename_button.addEventListener("click", () => {
				const new_name = window.prompt(`Enter a new name for '${entry.name}'.`, entry.name);
				if (new_name === null || new_name.trim() === "") return;
				console.log(new_name);
				fetch(getFetchBall('rename-file', `path=${encodeURIComponent(entry["relative-path"].replaceAll("/", "~/~"))}&newname=${encodeURIComponent(new_name)}`))
				.then(response => {
					if (response.ok) get_files(sessionStorage.getItem("current_dir"));
				});
				contextMenu.style.display = "";
				contextMenuButton.style.backgroundColor = ""
			});
			download_button.addEventListener("click", () => {
				const link = document.createElement("a");
				link.style.display = "none";
				link.href = `/yrFiles/files/${encodeURI(entry["relative-path"])}`;
				link.target = "_blank";
				link.download = entry.name;
				link.click();
				link.remove();
				contextMenu.style.display = "";
				contextMenuButton.style.backgroundColor = ""
			});
			edit_button.addEventListener("click", () => {
				const a = document.createElement("a");
				a.href = getFetchBall("yrText/", `relative-path=${encodeURI(entry["relative-path"])}`);
				a.target = "_blank";
				document.body.appendChild(a);
				a.click();
				a.remove();
				contextMenu.style.display = "";
				contextMenuButton.style.backgroundColor = ""
			});
			contextMenuButton.addEventListener("click", () => {
				if (getComputedStyle(contextMenu).display == "none") {
					for (const cm of Array.from(files.getElementsByClassName("options"))) {
						cm.style.display = "";
						cm.parentElement.querySelector("button")
							.style.backgroundColor = "";
					}
					contextMenu.style.display = "flex";
					contextMenuButton.style.backgroundColor = "#444";
					const y = contextMenuButton.getBoundingClientRect().top + window.scrollY;
					contextMenu.style.top = Math.min(y, 485) + "px";
					contextMenu.style.left = `${contextMenuButton.getBoundingClientRect().left - 205}px`;
				} else {
					contextMenu.style.display = "";
					contextMenuButton.style.backgroundColor = ""
				}
			});
			entry_div.appendChild(entry_span);
			entry_div.appendChild(contextMenuButton);
			entry_div.appendChild(contextMenu);
			contextMenu.appendChild(delete_button);
			contextMenu.appendChild(rename_button);
			if (entry.type == "f") contextMenu.appendChild(download_button);
			const editable = (response) => {
				if (!response.ok) return false;
				const type = response.headers.get("Content-Type").split(";")[0].trim();
				const contentTypes = [
					"application/json",
					"application/javascript",
					"application/xml",
					"application/xhtml+xml",
					"application/ld+json",
					"application/graphql",
					"application/rss+xml",
					"application/atom+xml",
					"application/sql"
				];
				if (type.startsWith("text/") || contentTypes.includes(type)) {
					return true;
				}
				return false;
			};

			if (entry.type == "f") fetch(`/yrFiles/files/${encodeURI(entry["relative-path"])}?username=root&password=root`, {method: "HEAD"})
			.then(response => {
				if (editable(response)) contextMenu.appendChild(edit_button);
			});
			files.appendChild(entry_div);
		}
	})
	.catch(error => {
		console.error('There was a problem with the fetch operation:', error);
	});
}
