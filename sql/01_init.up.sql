CREATE TABLE `account` (
   `username` varchar(32) NOT NULL,
   `password` varchar(32) NOT NULL,
   `level` tinyint(3) unsigned NOT NULL DEFAULT 0,
   `warn` tinyint(1) unsigned DEFAULT 0,
   `warn_expiration` int(11) unsigned DEFAULT 0,
   `suspended` tinyint(1) unsigned DEFAULT 0,
   `banned` tinyint(1) unsigned DEFAULT 0,
   `last_selected` int(11) DEFAULT NULL,
   `last_released` int(11) DEFAULT NULL,
   PRIMARY KEY (`username`)
);