-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Dec 11, 2025 at 04:00 PM
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

--
-- Dumping data for table `accounts`
--

INSERT INTO `accounts` (`id`, `account_number`, `account_name`, `account_type`, `building_id`, `isDefault`, `created_at`, `updated_at`) VALUES
(1, 1001, 'Account Receivable updated', 1, 2, 0, '2025-12-07 08:24:06', '2025-12-07 08:27:26'),
(2, 1001, 'Account Receivable', 1, 2, 0, '2025-12-07 08:24:17', '2025-12-07 08:24:17'),
(3, 1002, 'CASH ON HAND', 1, 2, 0, '2025-12-07 15:09:26', '2025-12-07 15:09:26'),
(4, 100, 'CASH ON HAND', 1, 1, 0, '2025-12-08 11:43:24', '2025-12-08 11:43:24'),
(5, 1003, 'DAHABSHIIL BANK', 1, 2, 0, '2025-12-08 13:29:30', '2025-12-08 13:29:30'),
(6, 1003, 'DAHABSHIIL BANK', 1, 1, 0, '2025-12-08 13:29:53', '2025-12-08 13:29:53'),
(7, 4000, 'MONTHLY RENT INCOME', 9, 1, 0, '2025-12-11 10:03:24', '2025-12-11 10:03:24'),
(8, 4001, 'SERVICE INCOME', 9, 1, 0, '2025-12-11 10:03:40', '2025-12-11 10:03:40'),
(9, 1010, 'ACCOUNT RECEIVABLE', 2, 1, 0, '2025-12-11 11:20:08', '2025-12-11 11:20:08'),
(10, 4100, 'DISCOUNT GIVEN', 9, 1, 0, '2025-12-11 11:26:01', '2025-12-11 11:26:37'),
(11, 2000, 'DEFFERED INCOME', 7, 1, 0, '2025-12-11 14:16:59', '2025-12-11 14:16:59');

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

--
-- Dumping data for table `account_types`
--

INSERT INTO `account_types` (`id`, `typeName`, `type`, `sub_type`, `typeStatus`, `created_at`, `updated_at`) VALUES
(1, 'Bank', 'Asset', 'current asset', 'debit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(2, 'Account Receivable', 'Asset', 'current asset', 'debit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(3, 'Other current asset', 'Asset', 'current asset', 'debit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(4, 'Cost of goods sold', 'Expense', '', 'debit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(5, 'Expense', 'Expense', '', 'debit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(6, 'Account Payable', 'Liability', '', 'credit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(7, 'Other Liability', 'Liability', '', 'credit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(8, 'Equity', 'Equity', '', 'credit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(9, 'Income', 'Income', '', 'credit', '2025-12-07 05:10:03', '2025-12-07 05:10:03'),
(10, 'Fixed Asset', 'Asset', 'fixed asset', 'debit', '2025-12-07 05:10:03', '2025-12-07 05:10:03');

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
(2, 'Hormuud', '2025-12-06 16:31:10', '2025-12-06 17:18:17'),
(3, 'Maansoor', '2025-12-08 07:13:27', '2025-12-08 07:14:30');

-- --------------------------------------------------------

--
-- Table structure for table `invoices`
--

CREATE TABLE `invoices` (
  `id` int(11) NOT NULL,
  `invoice_no` int(11) NOT NULL,
  `transaction_id` int(11) NOT NULL,
  `sales_date` date NOT NULL,
  `due_date` date NOT NULL,
  `ar_account_id` int(11) NOT NULL,
  `unit_id` int(11) DEFAULT NULL,
  `people_id` int(11) DEFAULT NULL,
  `user_id` int(11) NOT NULL,
  `amount` decimal(10,2) NOT NULL,
  `description` text NOT NULL,
  `refrence` varchar(50) NOT NULL,
  `cancel_reason` text DEFAULT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `building_id` int(11) NOT NULL,
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `invoices`
--

INSERT INTO `invoices` (`id`, `invoice_no`, `transaction_id`, `sales_date`, `due_date`, `ar_account_id`, `unit_id`, `people_id`, `user_id`, `amount`, `description`, `refrence`, `cancel_reason`, `status`, `building_id`, `createdAt`, `updatedAt`) VALUES
(6, 1, 6, '2025-12-11', '2026-01-10', 9, 1, 1, 1, 300.00, 'testing', 'test', NULL, '1', 1, '2025-12-11 10:08:36', '2025-12-11 10:52:32'),
(7, 2, 7, '2025-12-11', '2026-01-10', 9, 4, 2, 1, 140.00, 's', 's', NULL, '1', 1, '2025-12-11 10:14:12', '2025-12-11 10:52:34'),
(8, 4, 8, '2025-12-11', '2026-01-10', 9, 1, 2, 1, 350.00, 'a', 'a', NULL, '1', 1, '2025-12-11 10:18:23', '2025-12-11 13:01:23'),
(9, 5, 9, '2025-12-11', '2026-01-10', 9, 1, 1, 1, 255.00, 'df', 'f', NULL, '1', 1, '2025-12-11 10:19:40', '2025-12-11 13:01:23'),
(10, 6, 10, '2025-12-11', '2026-01-10', 9, 4, 3, 1, 300.00, 'a', 'a', NULL, '1', 1, '2025-12-11 10:23:44', '2025-12-11 10:52:42'),
(11, 8, 11, '2025-12-11', '2026-01-10', 9, 4, 1, 1, 300.00, 'a', 'f', NULL, '1', 1, '2025-12-11 10:28:04', '2025-12-11 10:52:45'),
(16, 10, 16, '2025-12-11', '2026-01-10', 9, 1, 1, 1, 300.00, 'f', 'a', NULL, '1', 1, '2025-12-11 10:58:09', '2025-12-11 13:01:23'),
(17, 12, 18, '2025-12-11', '2026-01-10', 9, 1, 1, 1, 0.00, 'd', 'd', NULL, '1', 1, '2025-12-11 11:32:13', '2025-12-11 13:01:23');

-- --------------------------------------------------------

--
-- Table structure for table `invoice_items`
--

CREATE TABLE `invoice_items` (
  `id` int(11) NOT NULL,
  `invoice_id` int(11) NOT NULL,
  `item_id` int(11) NOT NULL,
  `item_name` varchar(250) NOT NULL,
  `previous_value` decimal(10,3) DEFAULT NULL,
  `current_value` decimal(10,3) DEFAULT NULL,
  `qty` decimal(10,2) DEFAULT NULL,
  `rate` varchar(100) DEFAULT NULL,
  `total` decimal(10,2) NOT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `invoice_items`
--

INSERT INTO `invoice_items` (`id`, `invoice_id`, `item_id`, `item_name`, `previous_value`, `current_value`, `qty`, `rate`, `total`, `status`, `created_at`, `updated_at`) VALUES
(5, 6, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '300', 300.00, '1', '2025-12-11 10:08:36', '2025-12-11 10:09:39'),
(6, 7, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '300', 300.00, '1', '2025-12-11 10:14:12', '2025-12-11 10:14:44'),
(7, 7, 4, 'DISCOUNT TO CUSTOMER', NULL, 0.000, 1.00, '-10', 10.00, '1', '2025-12-11 10:14:12', '2025-12-11 10:14:48'),
(8, 7, 3, 'CASH PAYMENT - GURI BARWAQO', NULL, 0.000, 1.00, '-150', 150.00, '1', '2025-12-11 10:14:12', '2025-12-11 10:14:52'),
(9, 8, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '350', 350.00, '1', '2025-12-11 10:18:23', '2025-12-11 13:01:14'),
(10, 9, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '255', 255.00, '1', '2025-12-11 10:19:40', '2025-12-11 13:01:14'),
(11, 10, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '300', 300.00, '1', '2025-12-11 10:23:44', '2025-12-11 13:01:14'),
(12, 11, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '300', 300.00, '1', '2025-12-11 10:28:04', '2025-12-11 13:01:14'),
(13, 16, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '300', 300.00, '1', '2025-12-11 10:58:09', '2025-12-11 13:01:14'),
(14, 17, 1, 'MONTHLY RENT', NULL, 0.000, 1.00, '300', 300.00, '1', '2025-12-11 11:32:13', '2025-12-11 13:01:14'),
(15, 17, 6, 'RETAINER APPLIED', NULL, 0.000, 1.00, '-300', -300.00, '1', '2025-12-11 11:32:13', '2025-12-11 13:01:14');

-- --------------------------------------------------------

--
-- Table structure for table `invoice_payments`
--

CREATE TABLE `invoice_payments` (
  `id` int(11) NOT NULL,
  `transaction_id` int(11) NOT NULL,
  `date` date NOT NULL,
  `invoice_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `account_id` int(11) NOT NULL,
  `amount` decimal(10,2) NOT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `invoice_payments`
--

INSERT INTO `invoice_payments` (`id`, `transaction_id`, `date`, `invoice_id`, `user_id`, `account_id`, `amount`, `status`, `createdAt`, `updatedAt`) VALUES
(1, 19, '2025-12-11', 6, 1, 6, 300.00, '1', '2025-12-11 11:57:41', '2025-12-11 13:00:48');

-- --------------------------------------------------------

--
-- Table structure for table `items`
--

CREATE TABLE `items` (
  `id` int(11) NOT NULL,
  `name` varchar(250) NOT NULL,
  `type` enum('inventory','non inventory','service','discount','payment') NOT NULL,
  `description` text NOT NULL,
  `asset_account` int(11) DEFAULT NULL,
  `income_account` int(11) DEFAULT NULL,
  `cogs_account` int(11) DEFAULT NULL,
  `expense_account` int(11) DEFAULT NULL,
  `on_hand` decimal(10,2) NOT NULL,
  `avg_cost` decimal(10,2) NOT NULL,
  `date` date NOT NULL,
  `building_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `items`
--

INSERT INTO `items` (`id`, `name`, `type`, `description`, `asset_account`, `income_account`, `cogs_account`, `expense_account`, `on_hand`, `avg_cost`, `date`, `building_id`, `created_at`, `updated_at`) VALUES
(1, 'MONTHLY RENT', 'service', 'UNIT MONTHLY RENT', NULL, 7, NULL, NULL, 0.00, 0.00, '2025-12-11', 1, '2025-12-11 06:36:16', '2025-12-11 07:03:51'),
(2, 'SERVICE', 'service', 'WIISH', NULL, 8, NULL, NULL, 0.00, 0.00, '2025-12-11', 1, '2025-12-11 07:04:12', '2025-12-11 07:04:12'),
(3, 'CASH PAYMENT - GURI BARWAQO', 'payment', 'CASH PAYMENT', 4, NULL, NULL, NULL, 0.00, 0.00, '2025-12-11', 1, '2025-12-11 07:08:39', '2025-12-11 07:08:39'),
(4, 'DISCOUNT TO CUSTOMER', 'discount', '', NULL, 10, NULL, NULL, 0.00, 0.00, '2025-12-11', 1, '2025-12-11 08:26:22', '2025-12-11 08:26:22'),
(5, 'RETAINER DUE', 'service', '', NULL, 11, NULL, NULL, 0.00, 0.00, '2025-12-11', 1, '2025-12-11 11:17:21', '2025-12-11 11:17:21'),
(6, 'RETAINER APPLIED', 'service', '', NULL, 11, NULL, NULL, 0.00, 0.00, '2025-12-11', 1, '2025-12-11 11:17:48', '2025-12-11 11:17:48');

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
(1, 'Hudeifa abdirashid', '61111111111', 1, 1, '2025-12-08 08:33:33', '2025-12-08 08:33:33'),
(2, 'Khalid abdirashid', '61111111111', 2, 1, '2025-12-08 08:33:33', '2025-12-08 08:41:51'),
(3, '#2 Hudeifa abdirashi', '61111111111', 3, 1, '2025-12-08 08:38:02', '2025-12-08 08:38:02'),
(4, 'Hormuud Customer', '615882522', 1, 2, '2025-12-08 10:28:32', '2025-12-08 10:28:32');

-- --------------------------------------------------------

--
-- Table structure for table `people_types`
--

CREATE TABLE `people_types` (
  `id` int(11) NOT NULL,
  `title` varchar(50) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `people_types`
--

INSERT INTO `people_types` (`id`, `title`) VALUES
(1, 'customer'),
(2, 'vendor'),
(3, 'employee'),
(6, 'others');

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
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `periods`
--

INSERT INTO `periods` (`id`, `period_name`, `start`, `end`, `building_id`, `is_closed`, `created_at`, `updated_at`) VALUES
(1, 'BQ-FY2025', '2025-01-01', '2025-12-31', 2, 0, '2025-12-07 04:59:01', '2025-12-07 05:00:50'),
(2, 'FY-2025', '2025-01-01', '2025-12-31', 1, 0, '2025-12-08 08:43:06', '2025-12-08 08:43:06');

-- --------------------------------------------------------

--
-- Table structure for table `receipt_items`
--

CREATE TABLE `receipt_items` (
  `id` int(11) NOT NULL,
  `receipt_id` int(11) NOT NULL,
  `item_id` int(11) NOT NULL,
  `item_name` varchar(250) NOT NULL,
  `previous_value` decimal(10,3) DEFAULT NULL,
  `current_value` decimal(10,3) DEFAULT NULL,
  `qty` decimal(10,2) DEFAULT NULL,
  `rate` varchar(100) DEFAULT NULL,
  `total` decimal(10,2) NOT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `receipt_items`
--

INSERT INTO `receipt_items` (`id`, `receipt_id`, `item_id`, `item_name`, `previous_value`, `current_value`, `qty`, `rate`, `total`, `status`, `created_at`, `updated_at`) VALUES
(1, 1, 5, 'RETAINER DUE', NULL, 0.000, 1.00, '300', 300.00, '0', '2025-12-11 11:22:31', '2025-12-11 11:22:31');

-- --------------------------------------------------------

--
-- Table structure for table `sales_receipt`
--

CREATE TABLE `sales_receipt` (
  `id` int(11) NOT NULL,
  `receipt_no` int(11) NOT NULL,
  `transaction_id` int(11) NOT NULL,
  `receipt_date` date NOT NULL,
  `unit_id` int(11) DEFAULT NULL,
  `people_id` int(11) DEFAULT NULL,
  `user_id` int(11) NOT NULL,
  `account_id` int(11) NOT NULL,
  `amount` decimal(10,2) NOT NULL,
  `description` text DEFAULT NULL,
  `cancel_reason` text DEFAULT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `building_id` int(11) NOT NULL,
  `createdAt` timestamp NOT NULL DEFAULT current_timestamp(),
  `updatedAt` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `sales_receipt`
--

INSERT INTO `sales_receipt` (`id`, `receipt_no`, `transaction_id`, `receipt_date`, `unit_id`, `people_id`, `user_id`, `account_id`, `amount`, `description`, `cancel_reason`, `status`, `building_id`, `createdAt`, `updatedAt`) VALUES
(1, 1, 17, '2025-11-01', 1, 1, 1, 4, 300.00, 'A', NULL, '1', 1, '2025-12-11 11:22:31', '2025-12-11 13:00:18');

-- --------------------------------------------------------

--
-- Table structure for table `splits`
--

CREATE TABLE `splits` (
  `id` int(11) NOT NULL,
  `transaction_id` int(11) NOT NULL,
  `account_id` int(11) NOT NULL,
  `people_id` int(11) DEFAULT NULL,
  `debit` decimal(10,2) DEFAULT NULL,
  `credit` decimal(10,2) DEFAULT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `splits`
--

INSERT INTO `splits` (`id`, `transaction_id`, `account_id`, `people_id`, `debit`, `credit`, `status`, `created_at`, `updated_at`) VALUES
(2, 6, 9, 1, 300.00, NULL, '1', '2025-12-11 13:08:36', '2025-12-11 13:09:54'),
(3, 6, 7, 1, NULL, 300.00, '1', '2025-12-11 13:08:36', '2025-12-11 13:09:57'),
(4, 7, 9, 2, 140.00, NULL, '1', '2025-12-11 13:14:12', '2025-12-11 13:15:16'),
(5, 7, 10, 2, 10.00, NULL, '1', '2025-12-11 13:14:12', '2025-12-11 13:26:32'),
(6, 7, 4, 2, 150.00, NULL, '1', '2025-12-11 13:14:12', '2025-12-11 13:26:37'),
(7, 7, 7, 2, NULL, 300.00, '1', '2025-12-11 13:14:12', '2025-12-11 13:26:42'),
(8, 8, 9, 2, 350.00, NULL, '1', '2025-12-11 13:18:23', '2025-12-11 13:26:46'),
(9, 8, 7, 2, NULL, 350.00, '1', '2025-12-11 13:18:23', '2025-12-11 13:26:49'),
(10, 9, 9, 1, 255.00, NULL, '1', '2025-12-11 13:19:40', '2025-12-11 13:19:40'),
(11, 9, 7, 1, NULL, 255.00, '1', '2025-12-11 13:19:40', '2025-12-11 13:19:40'),
(12, 10, 9, 3, 300.00, NULL, '1', '2025-12-11 13:23:44', '2025-12-11 13:23:44'),
(13, 10, 7, 3, NULL, 300.00, '1', '2025-12-11 13:23:44', '2025-12-11 13:23:44'),
(14, 11, 9, 1, 300.00, NULL, '1', '2025-12-11 13:28:04', '2025-12-11 13:28:04'),
(15, 11, 7, 1, NULL, 300.00, '1', '2025-12-11 13:28:04', '2025-12-11 13:28:04'),
(16, 16, 9, 1, 300.00, NULL, '1', '2025-12-11 13:58:09', '2025-12-11 13:58:09'),
(17, 16, 7, 1, NULL, 300.00, '1', '2025-12-11 13:58:09', '2025-12-11 13:58:09'),
(18, 17, 4, 1, 300.00, NULL, '1', '2025-12-11 14:22:31', '2025-12-11 14:22:31'),
(19, 17, 11, 1, NULL, 300.00, '1', '2025-12-11 14:22:31', '2025-12-11 14:22:31'),
(20, 18, 7, 1, NULL, 300.00, '1', '2025-12-11 14:32:13', '2025-12-11 14:32:13'),
(21, 18, 11, 1, 300.00, NULL, '1', '2025-12-11 14:32:13', '2025-12-11 14:32:13'),
(22, 19, 6, 1, 300.00, NULL, '1', '2025-12-11 14:57:41', '2025-12-11 14:57:41'),
(23, 19, 9, 1, NULL, 300.00, '1', '2025-12-11 14:57:41', '2025-12-11 14:57:41');

-- --------------------------------------------------------

--
-- Table structure for table `transactions`
--

CREATE TABLE `transactions` (
  `id` int(11) NOT NULL,
  `type` enum('invoice','payment','check','deposit','bill','credit memo','sales receipt','journal','bill credit','bill payment') NOT NULL,
  `transaction_date` date NOT NULL,
  `memo` text NOT NULL,
  `status` enum('0','1') NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `building_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `unit_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `transactions`
--

INSERT INTO `transactions` (`id`, `type`, `transaction_date`, `memo`, `status`, `created_at`, `updated_at`, `building_id`, `user_id`, `unit_id`) VALUES
(6, 'invoice', '2025-12-11', 'testing', '1', '2025-12-11 13:08:36', '2025-12-11 15:59:59', 1, 1, 1),
(7, 'invoice', '2025-12-11', 's', '1', '2025-12-11 13:14:12', '2025-12-11 15:59:59', 1, 1, 4),
(8, 'invoice', '2025-12-11', 'a', '1', '2025-12-11 13:18:23', '2025-12-11 15:59:59', 1, 1, 1),
(9, 'invoice', '2025-12-11', 'df', '1', '2025-12-11 13:19:40', '2025-12-11 15:59:59', 1, 1, 1),
(10, 'invoice', '2025-12-11', 'a', '1', '2025-12-11 13:23:44', '2025-12-11 15:59:59', 1, 1, 4),
(11, 'invoice', '2025-12-11', 'a', '1', '2025-12-11 13:28:04', '2025-12-11 15:59:59', 1, 1, 4),
(16, 'invoice', '2025-12-11', 'f', '1', '2025-12-11 13:58:09', '2025-12-11 15:59:59', 1, 1, 1),
(17, 'sales receipt', '2025-11-01', 'A', '1', '2025-12-11 14:22:31', '2025-12-11 15:59:59', 1, 1, 1),
(18, 'invoice', '2025-12-11', 'd', '1', '2025-12-11 14:32:13', '2025-12-11 15:59:59', 1, 1, 1),
(19, '', '2025-12-11', 'Payment for Invoice #1', '1', '2025-12-11 14:57:41', '2025-12-11 15:59:59', 1, 1, 1);

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
(1, 'A1', 1, '2025-12-06 17:24:46', '2025-12-08 08:12:27'),
(2, 'A1', 3, '2025-12-08 07:13:57', '2025-12-08 07:13:57'),
(3, 'A2', 3, '2025-12-08 07:14:10', '2025-12-08 07:26:28'),
(4, 'A2', 1, '2025-12-08 08:09:40', '2025-12-08 08:12:17'),
(5, 'A1', 2, '2025-12-08 08:12:57', '2025-12-08 08:12:57');

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
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `name`, `username`, `phone`, `password`) VALUES
(1, 'Hudeifa abdirashid', 'hudeifa123', '615334355', '3333333333');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `accounts`
--
ALTER TABLE `accounts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `acc_acc_type_fk` (`account_type`),
  ADD KEY `accounts_building_fk` (`building_id`);

--
-- Indexes for table `account_types`
--
ALTER TABLE `account_types`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `buildings`
--
ALTER TABLE `buildings`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Indexes for table `invoices`
--
ALTER TABLE `invoices`
  ADD PRIMARY KEY (`id`),
  ADD KEY `invoice_unit_id` (`unit_id`),
  ADD KEY `invoice_people_id` (`people_id`),
  ADD KEY `invoice_user_id` (`user_id`),
  ADD KEY `invoice_building_id` (`building_id`),
  ADD KEY `fk_invoice_account_id` (`ar_account_id`);

--
-- Indexes for table `invoice_items`
--
ALTER TABLE `invoice_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `invoice_items_inv_id` (`invoice_id`),
  ADD KEY `invoice_items_item_id` (`item_id`);

--
-- Indexes for table `invoice_payments`
--
ALTER TABLE `invoice_payments`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ip_transaction_id` (`transaction_id`),
  ADD KEY `ip_invoice_id` (`invoice_id`),
  ADD KEY `ip_user_id` (`user_id`),
  ADD KEY `ip_account_id` (`account_id`);

--
-- Indexes for table `items`
--
ALTER TABLE `items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `item_asset_account_pk` (`asset_account`),
  ADD KEY `item_income_account_pk` (`income_account`),
  ADD KEY `item_cogs_account_pk` (`cogs_account`),
  ADD KEY `item_expense_account_pk` (`expense_account`),
  ADD KEY `item_building_fk` (`building_id`);

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
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `periods`
--
ALTER TABLE `periods`
  ADD PRIMARY KEY (`id`),
  ADD KEY `period_building_id` (`building_id`);

--
-- Indexes for table `receipt_items`
--
ALTER TABLE `receipt_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `sri_receipt_id` (`receipt_id`),
  ADD KEY `sri_item_id` (`item_id`);

--
-- Indexes for table `sales_receipt`
--
ALTER TABLE `sales_receipt`
  ADD PRIMARY KEY (`id`),
  ADD KEY `sr_unit_id` (`unit_id`),
  ADD KEY `sr_people_id` (`people_id`),
  ADD KEY `sr_user_id` (`user_id`),
  ADD KEY `sr_building_id` (`building_id`),
  ADD KEY `sr_account_id` (`account_id`);

--
-- Indexes for table `splits`
--
ALTER TABLE `splits`
  ADD PRIMARY KEY (`id`),
  ADD KEY `td_transaction_id` (`transaction_id`),
  ADD KEY `td_account_id` (`account_id`),
  ADD KEY `td_people_id` (`people_id`);

--
-- Indexes for table `transactions`
--
ALTER TABLE `transactions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `transaction_type_fk` (`type`),
  ADD KEY `transaction_building_fk` (`building_id`),
  ADD KEY `transaction_user_fk` (`user_id`),
  ADD KEY `transaction_unit_fk` (`unit_id`);

--
-- Indexes for table `units`
--
ALTER TABLE `units`
  ADD PRIMARY KEY (`id`),
  ADD KEY `unit_building_id_fk` (`building_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `accounts`
--
ALTER TABLE `accounts`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=12;

--
-- AUTO_INCREMENT for table `account_types`
--
ALTER TABLE `account_types`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT for table `buildings`
--
ALTER TABLE `buildings`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `invoices`
--
ALTER TABLE `invoices`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;

--
-- AUTO_INCREMENT for table `invoice_items`
--
ALTER TABLE `invoice_items`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=16;

--
-- AUTO_INCREMENT for table `invoice_payments`
--
ALTER TABLE `invoice_payments`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `items`
--
ALTER TABLE `items`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

--
-- AUTO_INCREMENT for table `people`
--
ALTER TABLE `people`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `people_types`
--
ALTER TABLE `people_types`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

--
-- AUTO_INCREMENT for table `periods`
--
ALTER TABLE `periods`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `receipt_items`
--
ALTER TABLE `receipt_items`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `sales_receipt`
--
ALTER TABLE `sales_receipt`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `splits`
--
ALTER TABLE `splits`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=24;

--
-- AUTO_INCREMENT for table `transactions`
--
ALTER TABLE `transactions`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=20;

--
-- AUTO_INCREMENT for table `units`
--
ALTER TABLE `units`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `accounts`
--
ALTER TABLE `accounts`
  ADD CONSTRAINT `acc_acc_type_fk` FOREIGN KEY (`account_type`) REFERENCES `account_types` (`id`),
  ADD CONSTRAINT `accounts_building_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);

--
-- Constraints for table `invoices`
--
ALTER TABLE `invoices`
  ADD CONSTRAINT `fk_invoice_account_id` FOREIGN KEY (`ar_account_id`) REFERENCES `accounts` (`id`),
  ADD CONSTRAINT `fk_invoice_building` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_invoice_people` FOREIGN KEY (`people_id`) REFERENCES `people` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_invoice_unit` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

--
-- Constraints for table `invoice_payments`
--
ALTER TABLE `invoice_payments`
  ADD CONSTRAINT `fk_ip_account` FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_ip_invoice` FOREIGN KEY (`invoice_id`) REFERENCES `invoices` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_ip_transaction` FOREIGN KEY (`transaction_id`) REFERENCES `transactions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_ip_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE CASCADE;

--
-- Constraints for table `items`
--
ALTER TABLE `items`
  ADD CONSTRAINT `fk_items_building` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `people`
--
ALTER TABLE `people`
  ADD CONSTRAINT `people_building_id_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`),
  ADD CONSTRAINT `people_type_id_fk` FOREIGN KEY (`type_id`) REFERENCES `people_types` (`id`);

--
-- Constraints for table `periods`
--
ALTER TABLE `periods`
  ADD CONSTRAINT `period_building_id` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);

--
-- Constraints for table `receipt_items`
--
ALTER TABLE `receipt_items`
  ADD CONSTRAINT `fk_sri_item` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_sri_receipt` FOREIGN KEY (`receipt_id`) REFERENCES `sales_receipt` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `sales_receipt`
--
ALTER TABLE `sales_receipt`
  ADD CONSTRAINT `fk_sr_account` FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_sr_building` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_sr_people` FOREIGN KEY (`people_id`) REFERENCES `people` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_sr_unit` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

--
-- Constraints for table `splits`
--
ALTER TABLE `splits`
  ADD CONSTRAINT `fk_splits_people` FOREIGN KEY (`people_id`) REFERENCES `people` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

--
-- Constraints for table `transactions`
--
ALTER TABLE `transactions`
  ADD CONSTRAINT `fk_transactions_building` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_transactions_unit` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_transactions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `units`
--
ALTER TABLE `units`
  ADD CONSTRAINT `unit_building_id_fk` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
