'use strict';

var gulp = require('gulp'),
	fs = require('fs');

var filesToMove = {
	'../../../node_modules/sierra-library/dist/sierra.min.css': '../../../../resources/static/css'
};

gulp.task('dist', function() {
	for(var key in filesToMove) {
		var src = __dirname + "/" + key;
		gulp.src(src)
		.pipe(gulp.dest(__dirname + "/" + filesToMove[key]));
	}
});
