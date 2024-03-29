CREATE TABLE `account`
(
    `username`        varchar(32) NOT NULL,
    `password`        varchar(32) NOT NULL,
    `level`           tinyint(3) unsigned NOT NULL DEFAULT 0,
    `warn`            tinyint(1) unsigned DEFAULT 0,
    `warn_expiration` int(11) unsigned DEFAULT 0,
    `suspended`       tinyint(1) unsigned DEFAULT 0,
    `banned`          tinyint(1) unsigned DEFAULT 0,
    `invalid`         tinyint(1) unsigned DEFAULT 0,
    `last_disabled`   int(11) DEFAULT NULL,
    `last_banned`     int(11) DEFAULT NULL,
    `last_suspended`  int(11) DEFAULT NULL,
    `last_selected`   int(11) DEFAULT NULL,
    `last_released`   int(11) DEFAULT NULL,
    PRIMARY KEY (`username`)
);

CREATE TABLE `area`
(
    `id`                         int(10) unsigned NOT NULL AUTO_INCREMENT,
    `name`                       varchar(255) NOT NULL,
    `geofence`                   mediumtext DEFAULT NULL,
    `pokemon_mode_workers`       int(10) unsigned NOT NULL DEFAULT 0,
    `pokemon_mode_route`         mediumtext DEFAULT NULL,
    `fort_mode_workers`          int(10) unsigned NOT NULL DEFAULT 0,
    `fort_mode_route`            mediumtext DEFAULT NULL,
    `enable_quests`              tinyint(1) NOT NULL DEFAULT 0,
    `quest_mode_workers`         int(10) unsigned NOT NULL DEFAULT 0,
    `quest_mode_route`           mediumtext DEFAULT NULL,
    `quest_mode_hours`           text DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_area_name` (`name`)
);

CREATE TABLE `quest_check`
(
    `area_id`   int(10) unsigned NOT NULL,
    `lat`       float NOT NULL,
    `lon`       float NOT NULL,
    `pokestops` text  NOT NULL,
    PRIMARY KEY (`lat`, `lon`),
    KEY         `area_id` (`area_id`),
    CONSTRAINT `quest_check_ix` FOREIGN KEY (`area_id`) REFERENCES `area` (`id`)
);

CREATE TABLE `stats_workers` (
    `datetime` datetime NOT NULL,
    `controller_worker` varchar(255) NOT NULL,
    `device_name` varchar(255) DEFAULT NULL,
    `loc_avg` float DEFAULT NULL,
    `loc_count` int(11) DEFAULT NULL,
    `loc_success` int(11) DEFAULT NULL,
    `mons_seen` int(11) DEFAULT NULL,
    `mons_enc` int(11) DEFAULT NULL,
    `stops` int(11) DEFAULT NULL,
    `quests` int(11) DEFAULT NULL,
    PRIMARY KEY (`datetime`,`controller_worker`)
);
