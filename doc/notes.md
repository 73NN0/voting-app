## 28 nov 2025

I'm starting the implementation of the service with a database schema. 
I realise some missfealures, first I need to hash the email and the password and store them on a separate database.
I don't need to hard link the mdp and the email in the user struct.

My design is poorly secure at this moment and not fully logic, for example, do I need to update the password if I update my user ? In my design yes, but I can fix by separate the password from the user domain.... At the end It will be the auth service which will be in charge of the emails and passwords nothing else.

I need to add a updated_at col for the user.
need to study this `LEFT JOIN`
```sql
db.QueryRowContext(ctx, `SELECT id, name, email FROM user WHERE id = ?`)
db.QueryRowContext(ctx, `SELECT password_hash FROM user_password WHERE user_id = ?`)

// ✅ Peut faire en 1 avec JOIN
db.QueryRowContext(ctx, `
    SELECT u.id, u.name, u.email, p.password_hash 
    FROM user u 
    LEFT JOIN user_password p ON u.id = p.user_id 
    WHERE u.id = ?
`)
```

I restrain myself to create immediately a repository pattern to abstract the database (today: sqlite)
I need to create an entity specialy for the model database ( remembering threedotslab ) User here is not the same struct from the front service handler...