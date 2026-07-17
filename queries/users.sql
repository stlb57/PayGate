INSERT INTO users (
    id,
    name,
    email,
    password_hash
)
VALUES (
    $1,
    $2,
    $3,
    $4
);