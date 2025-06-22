const upload_files = document.getElementById('upload-files');
const upload_folders = document.getElementById('upload-folders');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const send_size = 3 * 1024 * 1024;
const progress_bar = document.getElementById("progress-bar");
progress_bar.remove();
let use_auth = sessionStorage.getItem("auth") == ""? false:true;

import {get_files} from "../explorer/script.js";
const getFetchBall = (ballModel, notes) => {
	let resBall = `/${ballModel}`;
	use_auth = sessionStorage.getItem("auth") == ""? false:true;
	if (use_auth) resBall = `${resBall}?auth=${sessionStorage.getItem("auth")}`;
	if (typeof notes !== 'undefined') {
		if (use_auth) resBall = `${resBall}&${notes}`;
		else resBall = `${resBall}?${notes}`;
	}

	return resBall;
}

const upload = async(type) => {
	const files = type == "file" ? fileInput.files : folderInput.files;
	document.body.appendChild(progress_bar);
	progress_bar.style.display = "flex";
	document.getElementById("complete-bar").style.width = "0%";
	progress_bar.style.animation = "come-up 300ms ease-out 500ms forwards";

	for (const file of files) {
		const filename = type == "file"? encodeURIComponent(file.name)
			: encodeURIComponent(file.webkitRelativePath.split("/").join("~/~"));
		const raw_filename = type == "file"? file.name:file.webkitRelativePath;
		document.querySelector("#progress-bar > span").textContent = `Uploading '${raw_filename.length > 25 ? raw_filename.slice(0, 25)+"[...]" : raw_filename}'...`;
		const uploadURL = getFetchBall('upload-chunk', `filename=${filename}`);
		let sent = 0;
		while (sent < file.size) {
			const to_append = file.slice(sent, sent + send_size);
			sent += to_append.size;
			await fetch(uploadURL, {method: 'POST', body: to_append});	
			document.getElementById("complete-bar").style.width = `${sent / file.size * 100}%`;
		}
	}

	progress_bar.style.animation = "none";
	progress_bar.style.display = "flex";
	progress_bar.style.animation = "come-down 300ms ease-out 500ms forwards";
	setTimeout(() => {
		progress_bar.style.animation = "none";
		progress_bar.style.display = "flex";
		document.getElementById("complete-bar").style.width = "0%";
		progress_bar.remove();
	}, 800);
	get_files(sessionStorage.getItem("current_dir"));
}

upload_files.addEventListener('click', () => fileInput.click());
upload_folders.addEventListener('click', () => folderInput.click());
fileInput.addEventListener('change', () => upload("file"));
folderInput.addEventListener('change', upload);
