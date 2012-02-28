all:
	(cd views/client; hulk *.html > ../../public/static/js/templates.js)
