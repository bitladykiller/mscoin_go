-- This script creates the three logical databases used by the refactored
-- go-zero services. They intentionally stay separated because the new project
-- mirrors service ownership boundaries instead of sharing one monolithic schema.
CREATE DATABASE IF NOT EXISTS market CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS ucenter CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS exchange CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

