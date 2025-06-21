const upload_files = document.getElementById('upload-files');
const upload_folders = document.getElementById('upload-folders');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const max_size = 10 * 1024 * 1024 * 1024;

import {get_files} from "../explorer/script.js";

const upload = async(type) => {
	const files = type == "file" ? fileInput.files : folderInput.files;
	
	for (const file of files) {
		if (file.size > max_size) {
			window.alert("One of the files were too big (max size is 10gb): "+file.name);
		}
	}

	const zip = new JSZip();
	for (const file of files) {
		const arrayBuffer = await file.arrayBuffer();
		zip.file(file.webkitRelativePath || file.name, arrayBuffer);
	}

	const blob = await zip.generateAsync({type: 'blob', streamFiles: true});

	const auth = sessionStorage.getItem('auth');
	const Parent = encodeURIComponent(sessionStorage.getItem('current_dir').split('/').join('~/~'));
	const uploadUrl = Parent !== ""
	? `/upload-files?auth=${auth}&parent=${Parent}`
	: `/upload-files?auth=${auth}`;

	fetch(uploadUrl, {method: 'POST', headers: {"Content-Type": "application/octet-stream"}, body: blob})
	.then(response => {
		if (response.ok) get_files(sessionStorage.getItem('current_dir'));
		else alert('Upload failed...');
	});
}

upload_files.addEventListener('click', () => fileInput.click());
upload_folders.addEventListener('click', () => folderInput.click());
fileInput.addEventListener('change', () => upload("file"));
folderInput.addEventListener('change', upload);
