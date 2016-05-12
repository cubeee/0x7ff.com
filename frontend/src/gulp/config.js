'use strict';

module.exports = {
	sass: {
		src: './src/sass/**/*.scss',
		dest: '../resources/static/css',
		includePaths: [
			'./src/css/'
		]
	},
	js: {
		src: './src/js/**/*.js',
		dest: '../resources/static/js'
	},
	css: {
		dest: '../../../resources/static/css'
	}
}
