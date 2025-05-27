// authentication
let auth;
const authinput = document.getElementById('password-input');
const authenticateButton = document.getElementById('authenticateButton');

authenticateButton.addEventListener("click", () => {
	fetch(`/verify?auth=${authinput.value}`)
	.then(response => {
		if (response.ok) {
			auth = authinput.value;
			authinput.parentElement.style.display = 'none';
			document.getElementById('mainContent').style.display = 'block';
			get_files();
		} else {
			document.querySelector('#authentication > span').textContent = "Invalid password.";
			setTimeout(() => {
				document.querySelector('#authentication > span').textContent = "Enter your password to access files.";
			}, 1000);
		}
	})
});

// main content
const files = document.getElementById('files');
let current_dir;

files.getElementsByTagName('span')[0].addEventListener('click', () => {
	if (current_dir !== '') {
		let x = current_dir.split('/');
		x.pop(); x = x.join("/");

		get_files(x);
	}
});

const delete_func = () => {
	window.alert("alr i deleted it");
	get_files(current_dir);
}

const get_files = (dir) => {
	let reach_to;
	if (typeof dir == "undefined" || dir == '') {
		reach_to = `/get-files?auth=${auth}`;
		current_dir = ''
	} else {
		reach_to = `/get-files?auth=${auth}&dir=${encodeURIComponent(dir.toString())}`;
		current_dir = dir;
	}

	fetch(reach_to)
	.then(response => {
		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}
		return response.json();
	})
	.then(data => {
		for (const entry of Array.from(files.childNodes)) {
			if (entry.id !== 'gb') { entry.remove(); }
		}

		if (data == null) { return; }
		for (const entry of data) {
			const entry_div = document.createElement("div");
			entry_div.style.display = 'flex';
			entry_div.style.gap = '4px';

			const entry_span = document.createElement('span');
			entry_span.textContent = `${entry.name} (${entry.type})`;
			entry_span.style.color = 'black';
			entry_span.style.display = 'block';
			entry_span.style.cursor = 'pointer';
			const delete_button = document.createElement('button');
			delete_button.textContent = 'Delete';
			const rename_button = document.createElement('button');
			rename_button.textContent = 'Rename';

			entry_span.addEventListener('click', () => {
				console.log('clicked');

				if (current_dir == '' && entry_span.textContent.endsWith(" (d)")) {
					get_files(entry_span.textContent.slice(0, -4));
				} else if (entry_span.textContent.endsWith(' (d)')) {
					get_files(current_dir+"/"+entry_span.textContent.slice(0, -4));
				}
			});
			
			delete_button.addEventListener('click', () => {
				if (current_dir == '') {
					fetch(`/delete-file?auth=${auth}&name=${entry.name}`)
					.then(response => {
						if (response.status === 404) {
							window.alert('no such file lmao');
						}
						if (response.status === 503) {
							window.alert('didnt delete lmao');
						}
						if (response.status === 302) delete_func();
					})
				} else {
					const path = current_dir.replaceAll("/", "~%2F~");
					console.log(path);

					fetch(`/delete-file?auth=${auth}&name=${path}~%2F~${entry.name}`)
					.then(response => {
						if (response.status === 404) {
							window.alert('no such file lmao');
						}
						if (response.status === 503) {
							window.alert('didnt delete lmao');
						}
						if (response.status === 302) delete_func();
					})
				}
			});

			rename_button.addEventListener("click", () => {
				const new_name = window.prompt('new name is..?');
				if (new_name == '' || typeof new_name == 'undefined') return;

				if (current_dir == '' || typeof current_dir == 'undefined') {
					fetch(`/rename-file?auth=${auth}&path=${entry.name}&newname=${new_name}`)
					.then(response => {
						if (response.ok) {
							window.alert("alr i renamed it");
							get_files();
						}
					});
				} else {
					const path = current_dir.replaceAll("/", "~%2F~");
					console.log(path);

					fetch(`/rename-file?auth=${auth}&path=${path}~%2F~${entry.name}&newname=${new_name}`)
					.then(response => {
						if (response.ok) {
							window.alert("alr i renamed it");
							get_files(current_dir)
						}
					});
				}
			});

			entry_div.appendChild(entry_span);
			entry_div.appendChild(delete_button);
			entry_div.appendChild(rename_button);
			files.appendChild(entry_div);
		}
	})
	.catch(error => {
		console.error('There was a problem with the fetch operation:', error);
	});
}

