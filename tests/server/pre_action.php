<?php
if (isset($_COOKIE['globalcookie2'])) {
	echo "Hello " . $_COOKIE['globalcookie2'];
}
else {
	$c = time();
	setcookie('globalcookie2', 'alice', $c + 99999999, '/', 'localhost');
	echo "Hello nobody !";
}

