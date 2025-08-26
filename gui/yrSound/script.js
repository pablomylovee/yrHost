// authentication
const go_on = () => {
    username = usernameInput.value;
    password = passwordInput.value;
	document.getElementById("authentication").style.display = "none";
	document.getElementById("mainContent").style.display = "flex";
	showAllSongs();
};

let username, password;
const authenticationDiv = document.getElementById("authentication");
const usernameInput = document.getElementById('username-input');
const passwordInput = document.getElementById('password-input');
const authenticateButton = document.getElementById('authenticateButton');

authenticateButton.addEventListener("click", () => {
	fetch(`/isUser-qm?username=${encodeURIComponent(usernameInput.value)}&password=${encodeURIComponent(passwordInput.value)}`)
	.then(response => {
		if (response.ok) go_on();
		else {
			document.querySelector('#authentication > span').textContent = "Invalid login.";
			setTimeout(() => {
				document.querySelector('#authentication > span').textContent = "Login to see your music.";
			}, 1000);
		}
	})
});

// main
const getFetchBall = (ballModel, notes) => {
	let resBall = `/${ballModel}`;
	resBall = `${resBall}?username=${username}&password=${password}`;
	if (typeof notes !== 'undefined') resBall = `${resBall}&${notes}`;

	return resBall;
}

const shortenName = (s, maxLength) => {
	if (s.length > maxLength) return s.slice(0, maxLength)+"...";
	else return s;	
}

navigator.mediaSession.setActionHandler("play", () => document.getElementById("song").play());
navigator.mediaSession.setActionHandler("pause", () => document.getElementById("song").pause());
let queue = [];
let shuffeledQueue = [];

const playSong = async(songDiv, id, clearQueue) => {
	const cover = document.querySelector("#player > .simg-cont > img");
	const title = document.querySelector("#player > .song-info > b");
	const artist = document.querySelector("#player > .song-info > span");
	const songBar = document.getElementById("song-bar");
	if (clearQueue == true) queue.length = 0;

	fetch(getFetchBall("get-song-info", `id=${id}`)).then(response => {
		if (response.ok) return response.json();
	}).then(async(response) => {
		title.textContent = response.title;
		artist.textContent = response.artist;
		const res = await fetch(getFetchBall("get-cover", `id=${id}`), { method: "HEAD" });
		
		if ("mediaSession" in navigator) {
			navigator.mediaSession.metadata = new MediaMetadata({
				title: response.title,
				artist: response.artist,
				album: response.album,
				artwork: [
					{
						src: getFetchBall("get-cover", `id=${id}`),
						type: res.headers.get("Content-Type"),
					}
				],
			});
		}
	});
	(async() => {
		document.getElementById("song").src = getFetchBall("get-song-blob", `id=${id}`);
		document.getElementById("song").load();
		
		const onload = () => {
			songBar.min = "0"; songBar.value = "0";
			songBar.max = `${document.getElementById("song").duration}`;
			document.getElementById("song").play();
			document.getElementById("song").removeEventListener("loadedmetadata", onload);
		}
		document.getElementById("song").addEventListener("loadedmetadata", onload);
	})();
	(async() => {
		const response = await fetch(getFetchBall("get-cover", `id=${id}`), {method: "HEAD"});
        if (!response.ok) throw new Error(`${response.status}`)
        const rawIMG = new Image();
		rawIMG.src = cover.src = getFetchBall("get-cover", `id=${id}`);

		rawIMG.onload = () => {
			cover.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`;
			cover.style.width = "auto";
		}
	})().catch(() => {
		cover.src = "./vectors/song.svg";
		cover.style.aspectRatio = "1/1";
		cover.style.width = "32px";
	});

	currentSong = id;
	for (const element of Array.from(document.getElementById("songs").getElementsByClassName("song"))) {
		element.querySelector(".simg-cont > div").style.display = "";
		element.style.backgroundColor = "";
	}
	songDiv.querySelector(".simg-cont > div").style.display = "flex";
	songDiv.style.backgroundColor = "#444";
}

let albumsClickable = true;
let songsClickable = true;
let artistsClickable = true;
let currentSong;
const showAllAlbums = async() => {	
	if (!albumsClickable) return;
	for (const element of Array.from(document.getElementById("albums").childNodes)) element.remove();
	
	fetch(getFetchBall("get-albums"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(albums => {
		albumsClickable = false;
		const content = document.getElementById("albums");
		content.style.display = "flex";

		document.getElementById("songs").style.display = "none";
		document.getElementById("artists").style.display = "none";
		songsClickable = artistsClickable = true;
		document.getElementById("show-songs").style.backgroundColor = "";
		document.getElementById("show-albums").style.backgroundColor = "#424242";
		document.getElementById("show-artists").style.backgroundColor = "";

		for (const album of albums) {
			const albumDiv = document.createElement("div");
			albumDiv.classList.add("album");
			content.appendChild(albumDiv);

			const albumIMG = document.createElement("img");
			fetch(getFetchBall("get-album"), {method: "POST", body: JSON.stringify({
				album: album.title, "album-artist": album.artist
			})}).then(response => {
				if (response.ok) return response.json();
			}).then(response => {
				fetch(getFetchBall(`get-cover`, `id=${response[0].id}`)).then(response => {
					if (response.ok) return response.blob();
				}).then(response => {
					const rawIMG = new Image();
					albumIMG.src = rawIMG.src = URL.createObjectURL(response);

					rawIMG.onload = () => {
						albumIMG.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`
						albumDiv.appendChild(albumIMG);
					}
				}).catch(() => {
					albumIMG.src = "./vectors/song.svg";
					albumIMG.style.aspectRatio = "1/1";
					albumIMG.style.width = "64px";
					albumDiv.appendChild(albumIMG);
				});
			});
			const albumInfoDiv = document.createElement("div");
			albumDiv.appendChild(albumInfoDiv);

			const aiH3 = document.createElement("h3");
			aiH3.textContent = shortenName(album.title, 20);
			albumInfoDiv.appendChild(aiH3);

			const aiSpan = document.createElement("span");
			aiSpan.textContent = `${shortenName(album.artist, 22)} • ${album.year}`;
			albumInfoDiv.appendChild(aiSpan);

			albumDiv.addEventListener("mouseenter", () => {
				albumInfoDiv.style.bottom = "0";
			});
			albumDiv.addEventListener("mouseleave", () => {
				albumInfoDiv.style.bottom = "-60px";
			});
		}
	})
}

const showAllArtists = async() => {
	if (!artistsClickable) return;
	for (const element of Array.from(document.getElementById("artists").childNodes)) element.remove();
	
	fetch(getFetchBall("get-artists"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(artists => {
		artistsClickable = false;
		const content = document.getElementById("artists");
		content.style.display = "flex";

		document.getElementById("songs").style.display = "none";
		document.getElementById("albums").style.display = "none";
		albumsClickable = songsClickable = true;
		document.getElementById("show-songs").style.backgroundColor = "";
		document.getElementById("show-albums").style.backgroundColor = "";
		document.getElementById("show-artists").style.backgroundColor = "#424242";

		for (const artist of artists) {
			const artistDiv = document.createElement("div");
			artistDiv.classList.add("artist");
			content.appendChild(artistDiv);

			const picCont = document.createElement("div");
			artistDiv.appendChild(picCont);

			const img = document.createElement("img");
			img.src = "../../vectors/artist.svg";
			picCont.appendChild(img);
			fetch(getFetchBall("get-artist-picture", `name=${encodeURIComponent(artist)}`))
			.then(response => {
				if (response.ok) return response.blob();
			}).then(image => {
				if (typeof image !== "undefined" && image.length !== 0) {
					const rawIMG = new Image();
					img.src = rawIMG.src = URL.createObjectURL(image);
					rawIMG.onload = () => {
						img.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`;
						picCont.appendChild(img);
					}
				} else {
					img.src = "./vectors/artist.svg";
					img.style.aspectRatio = `1/1`;
					picCont.appendChild(img);
				}
			});

			const artistSpan = document.createElement("span");
			artistSpan.textContent = shortenName(artist, 20);
			artistDiv.appendChild(artistSpan);
		}
	})
}

const showAllSongs = async() => {
	if (!songsClickable) return;
	fetch(getFetchBall("get-songs"))
	.then(response => {
		if (response.ok) return response.json();
	}).then(songs => {
		songsClickable = false;
		const content = document.getElementById("songs");
		for (const element of Array.from(content.children)) element.remove();
		content.style.display = "flex";
	
		document.getElementById("artists").style.display = "none";
		document.getElementById("albums").style.display = "none";
		albumsClickable = artistsClickable = true;
		document.getElementById("show-songs").style.backgroundColor = "#424242";
		document.getElementById("show-albums").style.backgroundColor = "";
		document.getElementById("show-artists").style.backgroundColor = "";

		for (const song of songs) {
			const songDiv = document.createElement("div");
			songDiv.classList.add("song");
			songDiv.dataset.id = song.id;
			content.appendChild(songDiv);
			songDiv.addEventListener("click", () => {
				if (songDiv.dataset.id != currentSong) playSong(songDiv, song.id);
			});
			
			queue.push(song.id);

			const songIMGCont = document.createElement("div");
			songIMGCont.classList.add("simg-cont");
			songDiv.appendChild(songIMGCont);

			const songIMG = document.createElement("img");
			(async() => {
                const response = await fetch(getFetchBall("get-cover", `id=${song.id}`), {method: "HEAD"});
                if (!response.ok) throw new Error(`${response.status}`);
                const rawIMG = new Image();
				songIMG.src = rawIMG.src = getFetchBall("get-cover", `id=${song.id}`);

				rawIMG.onload = () => {
					songIMG.style.aspectRatio = `${rawIMG.naturalWidth} / ${rawIMG.naturalHeight}`;
					songIMGCont.appendChild(songIMG);
				}
			})().catch(() => {
				songIMG.src = "./vectors/song.svg";
				songIMG.style.aspectRatio = "1/1";
				songIMG.style.width = "20px";
				songIMGCont.appendChild(songIMG);
			});
			const playingIcon = document.createElement("div");
			songIMGCont.appendChild(playingIcon);
			const piIMG = document.createElement("img");
			piIMG.src = "./vectors/playing.svg";
			piIMG.width = "20"; playingIcon.appendChild(piIMG);
			playingIcon.style.display = songDiv.dataset.id == currentSong? "flex":"none";			
			songDiv.style.backgroundColor = songDiv.dataset.id == currentSong? "#444":"";			
		
			const songInfo = document.createElement("div");
			songInfo.classList.add("sinfo");
			songDiv.appendChild(songInfo);
		
			const songTitle = document.createElement("b");
			songTitle.textContent = song.title;
			songInfo.appendChild(songTitle);

			const songArtist = document.createElement("span");
			songArtist.textContent = song.artist;
			songInfo.appendChild(songArtist);
		}
		const shuffleFunc = (array) => {
			const new_array = Array.from(array);
			for (let i = new_array.length - 1; i > 0; i--) {
				const j = Math.floor(Math.random() * (i + 1));
				[new_array[i], new_array[j]] = [new_array[j], new_array[i]];
			}
			return new_array;
		}; shuffeledQueue = shuffleFunc(queue);
	});
} 

let sidebaropened = false;
const sidebar = document.getElementById("sidebar");
const button = document.getElementById("sidebar-button");
const icon = button.querySelector("img");

const hideSidebar = () => {
	sidebaropened = false;
	sidebar.style.transform = "";
	icon.src = "./vectors/hamburger-menu.svg";
	button.style.top = "";
	setTimeout(() => {
		sidebar.style.display = "";
	}, 400);
}
const showSidebar = () => {
	sidebaropened = true;
	sidebar.style.display = "flex";
	sidebar.style.transform = "translateY(-100%)";
	requestAnimationFrame(() => {
		requestAnimationFrame(() => {
			sidebar.style.transform = "translateY(0)";
		});
	});

	icon.src = "./vectors/close.svg";
	button.style.top = "360px";
}
let shuffeled = false;
const nextSong = () => {
	if (shuffeled == false) {
		const indexOf = queue.indexOf(currentSong, 0);
		if (indexOf == -1) return;
		let nindexOf = indexOf+1;
		if (nindexOf == -1 || typeof queue[nindexOf] == "undefined") nindexOf = 0;

		let tsongDiv;
		for (const songDiv of Array.from(document.getElementById("songs").getElementsByClassName("song"))) {
			if (songDiv.dataset.id == queue[nindexOf]) {
				tsongDiv = songDiv;
				break;
			}
		}

		playSong(tsongDiv, queue[nindexOf]);
	} else {
		const indexOf = shuffeledQueue.indexOf(currentSong, 0);
		if (indexOf == -1) return
		let nindexOf = indexOf + 1;
		if (nindexOf == -1 || typeof shuffeledQueue[nindexOf] == "undefined") nindexOf = 0;

		let tsongDiv;
		for (const songDiv of Array.from(document.getElementById("songs").getElementsByClassName("song"))) {
			if (songDiv.dataset.id == shuffeledQueue[nindexOf]) {
				tsongDiv = songDiv;
				break;
			}
		}

		playSong(tsongDiv, shuffeledQueue[nindexOf]);
	}
}
const prevSong = () => {
	if (shuffeled == false) {
		const indexOf = queue.indexOf(currentSong, 0);
		if (indexOf == -1) return
		let nindexOf = indexOf-1;
		if (nindexOf == -1 || typeof queue[nindexOf] == "undefined") nindexOf = queue.length - 1;

		let tsongDiv;
		for (const songDiv of Array.from(document.getElementById("songs").getElementsByClassName("song"))) {
			if (songDiv.dataset.id == queue[nindexOf]) {
				tsongDiv = songDiv;
				break;
			}
		}

		playSong(tsongDiv, queue[nindexOf]);
	} else {
		const indexOf = shuffeledQueue.indexOf(currentSong, 0);
		if (indexOf == -1) return
		let nindexOf = indexOf - 1;
		if (nindexOf == -1 || typeof shuffeledQueue[nindexOf] == "undefined") nindexOf = shuffeledQueue.length - 1;

		let tsongDiv;
		for (const songDiv of Array.from(document.getElementById("songs").getElementsByClassName("song"))) {
			if (songDiv.dataset.id == shuffeledQueue[nindexOf]) {
				tsongDiv = songDiv;
				break;
			}
		}

		playSong(tsongDiv, shuffeledQueue[nindexOf]);
	}
}
button.addEventListener("click", () => sidebaropened? hideSidebar():showSidebar());
document.getElementById("show-songs").addEventListener("click", () => {
	hideSidebar();
	showAllSongs();
});
document.getElementById("show-artists").addEventListener("click", () => {
	hideSidebar();
	showAllArtists();
});
document.getElementById("show-albums").addEventListener("click", () => {
	hideSidebar();
	showAllAlbums();
});
document.getElementById("song").addEventListener("loadedmetadata", () => {
	const songBar = document.getElementById("song-bar");
	songBar.max = document.getElementById("song").duration;
});

document.getElementById("song-bar").addEventListener("input", () => {
	const songBar = document.getElementById("song-bar");
	document.getElementById("song").currentTime = songBar.value;
});

document.getElementById("song").addEventListener("timeupdate", () => {
	const songBar = document.getElementById("song-bar");
	songBar.value = document.getElementById("song").currentTime;

	const elapsed = document.querySelector("#player > .song-info > .time > .elapsed");
	const minutes = Math.floor(document.getElementById("song").currentTime / 60);
	const seconds = Math.floor(document.getElementById("song").currentTime % 60).toString().padStart(2, "0");
	elapsed.textContent = `${minutes}:${seconds}`;

	const total = document.querySelector("#player > .song-info > .time > .total");
	const minutes1 = Math.floor(document.getElementById("song").duration / 60);
	const seconds1 = Math.floor(document.getElementById("song").duration % 60).toString().padStart(2, "0");
	total.textContent = `${minutes1}:${seconds1}`;
});

document.getElementById("song").addEventListener("play", () => {
	const play = document.getElementById("play");
	play.querySelector("img").src = "./vectors/pause.svg";
});
document.getElementById("song").addEventListener("pause", () => {
	const play = document.getElementById("play");
	play.querySelector("img").src = "./vectors/play.svg";
});
document.getElementById("song").addEventListener("ended", nextSong);
document.getElementById("repeat").addEventListener("click", () => {
	if (document.getElementById("song").loop) {
		document.getElementById("song").loop = false;
		document.getElementById("repeat").querySelector("img").src = "./vectors/repeat.svg"
	} else {
		document.getElementById("song").loop = true;
		document.getElementById("repeat").querySelector("img").src = "./vectors/repeat-toggled.svg"
	}
});

document.getElementById("play").addEventListener("click", () => {
	document.getElementById("song").paused?
		document.getElementById("song").play():
		document.getElementById("song").pause();
});

document.getElementById("next").addEventListener("click", nextSong);
document.getElementById("previous").addEventListener("click", prevSong);
document.getElementById("shuffle").addEventListener("click", () => {
	shuffeled = !shuffeled;
	shuffeled?
		document.getElementById("shuffle").querySelector("img").src = "./vectors/shuffle-toggled.svg":
		document.getElementById("shuffle").querySelector("img").src = "./vectors/shuffle.svg";
});
navigator.mediaSession.setActionHandler("nexttrack", () => nextSong());
navigator.mediaSession.setActionHandler("previoustrack", () => prevSong());
navigator.mediaSession.setActionHandler('seekto', (details) => {
	if (details.seekTime !== undefined) {
		document.getElementById("song").currentTime = details.seekTime;
	}
});
