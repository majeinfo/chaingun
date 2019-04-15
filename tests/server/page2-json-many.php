<?php
	$data = array(
		'people' => array(
			array('name' => $_COOKIE['globalcookie']),
			array('name' => 'fred'),
			array('name' => 'tom'),
			array('name' => 'alicia'),
			array('name' => 'jean'),
			array('name' => 'roger'),
			array('name' => 'tim'),
			array('name' => 'james'),
			array('name' => 'paul'),
			array('name' => 'john'),
			array('name' => 'ringo'),
			array('name' => 'keith'),
			array('name' => 'pam'),
			array('name' => 'kurt'),
			array('name' => 'vince'),
			array('name' => 'scarlett'),
			array('name' => 'ewan'),
			array('name' => 'sylvester'),
			array('name' => 'natalie'),
			array('name' => 'jimmy'),
			array('name' => 'ava'),
			array('name' => 'george'),
		)
	);
	echo json_encode($data);

