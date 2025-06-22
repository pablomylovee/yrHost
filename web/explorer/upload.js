const upload_files = document.getElementById('upload-files');
const upload_folders = document.getElementById('upload-folders');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const send_size = 3 * 1024 * 1024;
const auth = sessionStorage.getItem("auth")
const use_auth = auth == ""? false:true

import {get_files} from "../explorer/script.js";
const getFetchBall = (ballModel, notes) => {
	let resBall = `/${ballModel}`;
	if (use_auth) resBall = `${resBall}?auth=${auth}`;
	if (typeof notes !== 'undefined') {
		if (use_auth) resBall = `${resBall}&${notes}`;
		else resBall = `${resBall}?${notes}`;
	}

	return resBall;
}

const upload = async(type) => {
	const files = type == "file" ? fileInput.files : folderInput.files;

	for (const file of files) {
		const use_path = type == "file" ? encodeURIComponent(file.name)
			: encodeURIComponent(file.webkitRelativePath.split("/").join("~/~"))
		const uploadURL = getFetchBall("upload-chunk", `filename=${use_path}`);
		let sent = 0;
		let chunks = [];
		while (sent < file.size) {
			const to_append = file.slice(sent, sent+send_size);
			chunks.push(to_append);
			sent += to_append.size;
		}

		fetch(uploadURL, {method: 'POST', body: file})
		.then(response => {
			if (response.ok) get_files(sessionStorage.getItem("current_dir"));
			else alert("Upload failed...");
		})
	}
}

upload_files.addEventListener('click', () => fileInput.click());
upload_folders.addEventListener('click', () => folderInput.click());
fileInput.addEventListener('change', () => upload("file"));
folderInput.addEventListener('change', upload);
