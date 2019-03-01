DROP DATABASE IF EXISTS {dbname}; 
CREATE DATABASE {dbname} CHARACTER SET = 'utf8mb4' COLLATE = 'utf8mb4_general_ci';

USE {dbname};

	CREATE TABLE Participant (
		participant_id INT AUTO_INCREMENT UNIQUE,
		first_name TEXT,
		`name` TEXT,
		email TEXT,
		street_name TEXT,
		street_number INT,
		postal_code INT,
		country TEXT,
		need_transport TINYINT(1),
		`comment` TEXT,
		PRIMARY KEY (participant_id),
		INDEX `idx_participant_name` (`name`(100)),
		INDEX `idx_participant_firstname` (`first_name`(100)),
		INDEX `idx_participant_country` (`country`(100))
	);

CREATE TABLE `Event` (
	event_id INT AUTO_INCREMENT UNIQUE,
	`name` TEXT,
	description TEXT,
	location TEXT,
	`date` DATETIME,
	start_time DATETIME,
	end_time DATETIME,
	max_participants INT,
	price DECIMAL,
	`comment` TEXT,
	PRIMARY KEY (event_id),
	INDEX `idx_event_name` (`name`(100))
);

CREATE TABLE Topic (
	topic_id INT AUTO_INCREMENT UNIQUE,
	event_id INT NOT NULL,
	`name` TEXT,
	description TEXT,
	referent TEXT,
	PRIMARY KEY (topic_id),
	CONSTRAINT `fk_topic_event` FOREIGN KEY (event_id)
		REFERENCES `Event` (`event_id`) 
			ON DELETE CASCADE
			ON UPDATE RESTRICT ,
	INDEX `idx_topic_name` (`name`(100))
);

CREATE TABLE Participant_Event (
	participant_id INT NOT NULL,
	event_id INT NOT NULL,
	PRIMARY KEY (`participant_id`, `event_id`),
	CONSTRAINT `fk_partevent_participant` FOREIGN KEY (`participant_id`)
		REFERENCES `Participant` (`participant_id`)
		ON DELETE CASCADE
		ON UPDATE RESTRICT,
	CONSTRAINT `fk_partevent_event` FOREIGN KEY (`event_id`)
		REFERENCES `Event` (`event_id`)
			ON DELETE CASCADE
			ON UPDATE RESTRICT
);
