<?php
if (isset($_COOKIE['globalcookie'])) {
	echo "Hello " . $_COOKIE['globalcookie'];
}
else {
	$c = time();
	setcookie('globalcookie', 'bob', $c + 99999999, '/', 'localhost');
	echo "Hello nobody !";
}

