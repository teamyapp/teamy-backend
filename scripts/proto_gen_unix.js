#!/usr/bin/env node

const fs = require('fs');
const {execSync} = require("child_process");
const paths = ['core', 'apps'];

for (let path of paths) {
	findFilesRec(path, (file) => {
		if (!file.endsWith('.proto')) {
			return;
		}

		execSync(`
		protoc --go_out=. --go_opt=paths=source_relative \\
		--go-grpc_out=. --go-grpc_opt=paths=source_relative ${file}`,
			{stdio: 'inherit'});
	});
}

function findFilesRec(dir, outputFile) {
	const files = fs.readdirSync(dir);
	files.forEach(file => {
		const path = `${dir}/${file}`;
		if (fs.statSync(path).isDirectory()) {
			findFilesRec(path, outputFile);
		} else {
			outputFile(path);
		}
	});
}


