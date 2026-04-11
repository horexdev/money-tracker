-- +goose Up

-- Step 1: Rename column
ALTER TABLE categories RENAME COLUMN emoji TO icon;

-- Step 2: Convert known emoji values to Phosphor icon IDs
UPDATE categories SET icon = CASE icon
    -- Food
    WHEN '🍔' THEN 'fork-knife'
    WHEN '🍕' THEN 'pizza'
    WHEN '🍳' THEN 'fork-knife'  WHEN '🥗' THEN 'fork-knife'
    WHEN '🍜' THEN 'fork-knife'  WHEN '🍱' THEN 'fork-knife'  WHEN '🥘' THEN 'fork-knife'
    -- Transport
    WHEN '🚕' THEN 'taxi'        WHEN '🚗' THEN 'car'
    WHEN '🚌' THEN 'bus'         WHEN '🚎' THEN 'bus'
    WHEN '🏍️' THEN 'car'        WHEN '🏍' THEN 'car'
    WHEN '🚂' THEN 'car'         WHEN '⛽' THEN 'car'
    -- Entertainment
    WHEN '🎬' THEN 'film-slate'  WHEN '🎭' THEN 'film-slate'  WHEN '🎪' THEN 'film-slate'
    WHEN '🎮' THEN 'game-controller'
    WHEN '🎵' THEN 'music-note'  WHEN '🎶' THEN 'music-note'
    -- Shopping
    WHEN '🛍️' THEN 'shopping-bag' WHEN '🛍' THEN 'shopping-bag' WHEN '🛒' THEN 'shopping-bag'
    -- Health
    WHEN '💊' THEN 'first-aid'   WHEN '🏥' THEN 'first-aid'   WHEN '🩺' THEN 'first-aid'
    -- Money / income
    WHEN '💰' THEN 'money'       WHEN '💵' THEN 'money'       WHEN '💸' THEN 'money'
    WHEN '💲' THEN 'currency-dollar' WHEN '💼' THEN 'briefcase' WHEN '🧾' THEN 'money'
    -- Tech
    WHEN '💻' THEN 'laptop'      WHEN '🖥️' THEN 'laptop'
    WHEN '📱' THEN 'device-mobile' WHEN '📲' THEN 'device-mobile'
    -- Home
    WHEN '🏠' THEN 'house'       WHEN '🏡' THEN 'house'
    WHEN '🛋️' THEN 'bed'        WHEN '🛋' THEN 'bed'
    -- Other
    WHEN '📦' THEN 'tag'         WHEN '📫' THEN 'tag'         WHEN '📬' THEN 'tag'
    WHEN '☕' THEN 'coffee'      WHEN '⚡' THEN 'lightning'    WHEN '❤️' THEN 'heartbeat'
    WHEN '🎓' THEN 'graduation-cap' WHEN '✈️' THEN 'airplane' WHEN '🎁' THEN 'gift'
    WHEN '🐾' THEN 'paw-print'   WHEN '🐶' THEN 'paw-print'  WHEN '🐱' THEN 'paw-print'
    WHEN '👕' THEN 't-shirt'     WHEN '💪' THEN 'barbell'     WHEN '👶' THEN 'baby'
    WHEN '🌸' THEN 'flower'      WHEN '🌺' THEN 'flower'
    WHEN '📚' THEN 'book-open'   WHEN '📖' THEN 'book-open'
    WHEN '🔧' THEN 'wrench'      WHEN '✂️' THEN 'scissors'
    WHEN '📷' THEN 'camera'      WHEN '📸' THEN 'camera'
    WHEN '🔒' THEN 'shield-check' WHEN '👛' THEN 'wallet'
    -- Banking
    WHEN '🏦' THEN 'piggy-bank'
    -- System: Transfer
    WHEN '↔️' THEN 'arrows-left-right' WHEN '↔' THEN 'arrows-left-right'
    -- System: Adjustment
    WHEN '⚖️' THEN 'scales'     WHEN '⚖' THEN 'scales'
    ELSE icon
END
WHERE icon !~ '^[a-z]';

-- Step 3: Catch-all for any unknown emoji
UPDATE categories SET icon = 'star' WHERE icon !~ '^[a-z]';

-- Step 4: Update default
ALTER TABLE categories ALTER COLUMN icon SET DEFAULT 'star';

-- +goose Down
ALTER TABLE categories ALTER COLUMN icon SET DEFAULT '';
UPDATE categories SET icon = '' WHERE icon ~ '^[a-z]';
ALTER TABLE categories RENAME COLUMN icon TO emoji;
