const upload_files = document.getElementById('upload-files');
const upload_folders = document.getElementById('upload-folders');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const send_size = 3 * 1024 * 1024;
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

	for (const file of files) {
		const filename = type == "file"? encodeURIComponent(file.name)
			: encodeURIComponent(file.webkitRelativePath.split("/").join("~/~"));
		const uploadURL = getFetchBall('upload-chunk', `filename=${filename}`)
		let sent = 0;
		while (sent < file.size) {
			const to_append = file.slice(sent, sent + send_size);
			sent += to_append.size;
			await fetch(uploadURL, {method: 'POST', body: to_append});	
		}
	}

	get_files(sessionStorage.getItem("current_dir"));
}

upload_files.addEventListener('click', () => fileInput.click());
upload_folders.addEventListener('click', () => folderInput.click());
fileInput.addEventListener('change', () => upload("file"));
folderInput.addEventListener('change', upload);
