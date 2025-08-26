const upload_files = document.getElementById('upload-files');
const upload_folders = document.getElementById('upload-folders');
const fileInput = document.getElementById('fileInput');
const folderInput = document.getElementById('folderInput');
const send_size = 3 * 1024 * 1024;
const progress_bar = document.getElementById("progress-bar");
progress_bar.remove();

import {get_files, getFetchBall} from "./script.js";
const usedNames = [];

const getRandomID = () => {
    let id = Math.floor(Math.random() * (999999 - 100000) + 100000);
    while (usedNames.indexOf(id) != -1)
        id = Math.floor(Math.random() * (999999 - 100000) + 100000);
    usedNames.push(id.valueOf());
    return id.valueOf();
}

const upload = async(type) => {
    upload_files.removeEventListener("click", ufic);
    upload_folders.removeEventListener("click", ufoc);

    const files = type == "file" ? fileInput.files : folderInput.files;
    const packID = getRandomID();

    document.body.appendChild(progress_bar);
    progress_bar.style.display = "flex";
    document.getElementById("complete-bar").style.width = "0%";
    progress_bar.style.animation = "come-up 300ms ease-out 500ms forwards";
    let bytesSent = 0;

    try {
        const r = await fetch(getFetchBall("create-pack", `service=files&name=${packID}`));
        if (!r.ok) return;

        let usedEntryIDs = [];

        let responses = [];
        for (const file of files) {
            let chunksSent = 0;
            const raw_filename = type == "file"? file.name:file.webkitRelativePath;
            let entryID;
            do {
                entryID = Math.floor(Math.random() * 899999 + 100000);
            } while (usedEntryIDs.includes(entryID));
            usedEntryIDs.push(entryID);

            const re = await fetch(getFetchBall("create-entry", `service=files&name=${packID}&id=${file.name}`), {
                method: "POST",
                body: JSON.stringify({ "relative-path": raw_filename })
            });

            if (!re.ok) return;

            document.querySelector("#progress-bar > span").textContent =
                `Uploading '${raw_filename.length > 32 ? raw_filename.slice(0, 25).trim()+" [...]" : raw_filename}'...`;


            while (bytesSent < file.size) {
                const to_append = file.slice(bytesSent, bytesSent + send_size);
                if (responses.length % 5 == 0) {
                    await Promise.allSettled(responses);
                    responses.length = 0;
                }
                responses.push(
                    fetch(getFetchBall("append-chunk", `service=files&name=${packID}&id=${encodeURI(file.name)}&part=${chunksSent}`), {
                        method: 'POST',
                        body: to_append
                    }).then(response => {
                        if (response.ok) {
                            chunksSent++;
                            bytesSent += to_append.size;
                            document.getElementById("complete-bar").style.width = `${(bytesSent / Array.from(files).reduce((sum, f) => sum + f.size, 0)) * 100}%`;
                        }
                        return response;
                    })
                );
            }
        }

        document.getElementById("complete-bar").style.width = `0`;
        const res = await fetch(getFetchBall("assemble-pack", `service=files&name=${packID}`));
        if (!res.ok) return;
    } finally {
        progress_bar.style.animation = "come-down 300ms ease-out forwards";
        setTimeout(() => {
            progress_bar.style.animation = "none";
            progress_bar.remove();
        }, 800);

        upload_files.addEventListener("click", ufic);
        upload_folders.addEventListener("click", ufoc);
    }

    usedNames.splice(usedNames.indexOf(packID), 1);
    get_files(sessionStorage.getItem("current_dir"));
};

const ufic = () => fileInput.click();
const ufoc = () => folderInput.click();
upload_files.addEventListener('click', ufic);
upload_folders.addEventListener('click', ufoc);
fileInput.addEventListener('change', async() => {
	await upload('file'); fileInput.value = '';
});
folderInput.addEventListener('change', async() => {
	await upload(); folderInput.value = '';
});
