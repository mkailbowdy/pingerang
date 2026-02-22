CREATE TABLE `sites` (
  `id` int NOT NULL AUTO_INCREMENT,
  `url` varchar(2038) NOT NULL,
  `created` datetime NOT NULL,
  `urlhash` char(64) NOT NULL,
  `pagehash` char(64) NOT NULL,
  `selector` varchar(256) NOT NULL,
  `changed` tinyint(1) NOT NULL DEFAULT '0',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `urlhash` (`urlhash`),
  KEY `idx_sites_created` (`created`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `sessions` (
  `token` char(43) NOT NULL,
  `data` blob NOT NULL,
  `expiry` timestamp(6) NOT NULL,
  PRIMARY KEY (`token`),
  KEY `sessions_expiry_idx` (`expiry`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
