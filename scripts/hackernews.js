function root() {
	var result = [];
	result.push({type: "dir", name: "beststories", url: "https://hacker-news.firebaseio.com/v0/beststories.json"});
	result.push({type: "dir", name: "newstories", url: "https://hacker-news.firebaseio.com/v0/newstories.json"});
	result.push({type: "dir", name: "topstories", url: "https://hacker-news.firebaseio.com/v0/topstories.json"});
	return result;
}

function list(url) {
	var data = httpGet(url);
	var json = JSON.parse(data);
	var result = [];
	for (i = 0; i < 30; i++) {
		var elm = json[i];
		var file = "https://hacker-news.firebaseio.com/v0/item/" + elm + ".json"
		var story = JSON.parse(httpGet(file));
		result.push({type: "file", name: story.score.toString() + " " + story.title + ".txt", url: file, size: "4096"});
	}
	return result;
}

function download(url) {
	var data = httpGet(url);
	var json = JSON.parse(data);
	var result = "";
	var len = json.kids.length < 30 ? json.kids.length : 30;
	for (i = 0; i < len; i++) {
		var elm = json.kids[i];
		var file = "https://hacker-news.firebaseio.com/v0/item/" + elm + ".json"
		var comment = JSON.parse(httpGet(file));
		result += comment.text + "\n";
	}
	return result;
}
