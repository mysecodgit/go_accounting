-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Dec 06, 2025 at 07:32 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `go_accounting`
--

-- --------------------------------------------------------

--
-- Table structure for table `buildings`
--

CREATE TABLE `buildings` (
  `id` int(11) NOT NULL,
  `name` varchar(250) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `buildings`
--

INSERT INTO `buildings` (`id`, `name`, `created_at`, `updated_at`) VALUES
(1, 'Guri Barwaaqo', '2025-12-06 16:29:00', '2025-12-06 16:29:00'),
(2, 'Guri Nabad updated', '2025-12-06 16:31:10', '2025-12-06 17:18:17');

-- --------------------------------------------------------

--
-- Table structure for table `people`
--

CREATE TABLE `people` (
  `id` int(11) NOT NULL,
  `name` varchar(20) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `type_id` int(11) NOT NULL,
  `building_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `people`
--

INSERT INTO `people` (`id`, `name`, `phone`, `type_id`, `building_id`, `created_at`, `updated_at`) VALUES
(3, 'Hudeifa abdirashid', '617889987', 1, 1, '2025-12-06 18:09:24', '2025-12-06 18:09:24'),
(4, 'Hudeifa abdirashid 1', '617889987', 4, 2, '2025-12-06 18:26:06', '2025-12-06 18:31:45');

-- --------------------------------------------------------

--
-- Table structure for table `people_types`
--

CREATE TABLE `people_types` (
  `id` int(11) NOT NULL,
  `title` varchar(50) NOT NULL,
  `building_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `people_types`
--

INSERT INTO `people_types` (`id`, `title`, `building_id`) VALUES
(1, 'Customer', 2),
(2, 'vendor', 1),
(3, 'employee', 1),
(4, 'employee', 2);

-- --------------------------------------------------------

--
-- Table structure for table `units`
--

CREATE TABLE `units` (
  `id` int(11) NOT NULL,
  `name` varchar(20) NOT NULL,
  `building_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `units`
--

INSERT INTO `units` (`id`, `name`, `building_id`, `created_at`, `updated_at`) VALUES
(1, 'A1 updated', 1, '2025-12-06 17:24:46', '2025-12-06 17:25:15');

-- --------------------------------------------------------

--
-- Table structure for table `periods`
--

CREATE TABLE `periods` (
  `id` int(11) NOT NULL,
  `period_name` varchar(50) NOT NULL,
  `start` date NOT NULL,
  `end` date NOT NULL,
  `building_id` int(11) NOT NULL,
  `is_closed` int(2) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `account_types`
--

CREATE TABLE `account_types` (
  `id` int(11) NOT NULL,
  `typeName` varchar(250) NOT NULL,
  `type` varchar(20) NOT NULL,
  `sub_type` varchar(20) NOT NULL,
  `typeStatus` varchar(10) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `accounts`
--

CREATE TABLE `accounts` (
  `id` int(11) NOT NULL,
  `account_number` int(11) NOT NULL,
  `account_name` varchar(50) NOT NULL,
  `account_type` int(11) NOT NULL,
  `building_id` int(11) NOT NULL,
  `isDefault` tinyint(1) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `name` varchar(50) NOT NULL,
  `username` varchar(20) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `password` varchar(10) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `buildings`
--
ALTER TABLE `buildings`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Indexes for table `people`
--
ALTER TABLE `people`
  ADD PRIMARY KEY (`id`),
  ADD KEY `people_type_id_fk` (`type_id`),
  ADD KEY `people_building_id_fk` (`building_id`);

--
-- Indexes for table `people_types`
--
ALTER TABLE `people_types`
  ADD PRIMARY KEY (`id`),
  ADD KEY `pt_building_id_fk` (`building_id`);

--
-- Indexes for table `units`
--
ALTER TABLE `units`
  ADD PRIMARY KEY (`id`),
  ADD KEY `unit_building_id_fk` (`building_id`);

--
-- Indexes for table `periods`
--
ALTER TABLE `periods`
  ADD PRIMARY KEY (`id`),
  ADD KEY `period_building_id` (`building_id`),
  ADD UNIQUE KEY `unique_building_period` (`building_id`, `start`, `end`);

--
-- Indexes for table `account_types`
--
ALTER TABLE `account_types`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `accounts`
--
ALTER TABLE `accounts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `acc_acc_type_fk` (`account_type`),
  ADD KEY `accounts_building_fk` (`building_id`),
  ADD UNIQUE KEY `unique_building_account_number` (`building_id`, `account_number`),
  ADD UNIQUE KEY `unique_building_account_name` (`building_id`, `account_name`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `buildings`
--
ALTER TABLE `buildings`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `people`
--
ALTER TABLE `people`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `people_types`
--
ALTER TABLE `people_types`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `units`
--
ALTER TABLE `units`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `periods`
--
ALTER TABLE `periods`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `account_types`
--
ALTER TABLE `account_types`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `accounts`
--
ALTER TABLE `accounts`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `people`
--
ALTER TABLE `people`
  ADD CONSTRAINT `people_building_id_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`),
  ADD CONSTRAINT `people_type_id_fk` FOREIGN KEY (`type_id`) REFERENCES `people_types` (`id`);

--
-- Constraints for table `people_types`
--
ALTER TABLE `people_types`
  ADD CONSTRAINT `pt_building_id_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);

--
-- Constraints for table `units`
--
ALTER TABLE `units`
  ADD CONSTRAINT `unit_building_id_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);

--
-- Constraints for table `periods`
--
ALTER TABLE `periods`
  ADD CONSTRAINT `period_building_id_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);

--
-- Constraints for table `accounts`
--
ALTER TABLE `accounts`
  ADD CONSTRAINT `acc_acc_type_fk` FOREIGN KEY (`account_type`) REFERENCES `account_types` (`id`),
  ADD CONSTRAINT `accounts_building_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
