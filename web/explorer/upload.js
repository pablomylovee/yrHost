const upload_files = document.getElementById('upload-files');
const upload_folders = document.getElementById('upload-folders');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const max_size = 10 * 1024 * 1024 * 1024;

import {get_files} from "../explorer/script.js";

const uploadf = async() => {
	for (const file of fileInput.files) {
		if (file.size > max_size) {
			window.alert("one of em files were too big (max size is 10gb): "+file.name);
		}
	}

	const form_data = new FormData();
	for (const file of fileInput.files) {
		form_data.append('files[]', file);
	}

	const auth = sessionStorage.getItem("auth");
	const Parent = encodeURIComponent(sessionStorage.getItem("current_dir").split("/").join("~/~"));
	if (Parent !== "") {
		fetch(`/upload-files?auth=${auth}&parent=${Parent}`, {method: 'POST', body: form_data})
		.then(response => {
			if (response.ok) get_files(sessionStorage.getItem("current_dir"));
		});
	} else {
		fetch(`/upload-files?auth=${auth}`, {method: 'POST', body: form_data})
		.then(response => {
			if (response.ok) get_files(sessionStorage.getItem("current_dir"));
		});
	}
}
const uploadd = async() => {
	for (const file of folderInput.files) {
		if (file.size > max_size) {
			window.alert("one of em files were too big (max size is 10gb): "+file.name);
		}
	}

	const form_data = new FormData();
	for (const file of folderInput.files) {
		form_data.append('files[]', file);
		form_data.append('paths[]', file.webkitRelativePath);
	}

	const auth = sessionStorage.getItem("auth");
	const Parent = encodeURIComponent(sessionStorage.getItem("current_dir").split("/").join("~/~"));
	if (Parent !== "") {
		fetch(`/upload-folders?auth=${auth}&parent=${Parent}`, {method: 'POST', body: form_data})
		.then(response => {
			if (response.ok) get_files(sessionStorage.getItem("current_dir"));
		});
	} else {
		fetch(`/upload-folders?auth=${auth}`, {method: 'POST', body: form_data})
		.then(response => {
			if (response.ok) get_files(sessionStorage.getItem("current_dir"));
		});
	}
}

upload_files.addEventListener('click', () => fileInput.click());
upload_folders.addEventListener('click', () => folderInput.click());
fileInput.addEventListener('change', uploadf);
folderInput.addEventListener('change', uploadd);

