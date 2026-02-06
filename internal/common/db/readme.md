# db

## Schema 

```mermaid
	erDiagram
    user ||--o{ user_password : "has"
    user ||--o{ session_and_participant : "participates"
    user ||--o{ vote : "casts"
    user ||--o{ user_history : "has_receipt"
    
    vote_session ||--o{ session_and_participant : "has_participants"
    vote_session ||--o{ question : "contains"
    vote_session ||--o{ vote : "receives"
    vote_session ||--o{ user_history : "generates_receipts"
    vote_session ||--o{ result_history : "has_result"
    
    question ||--o{ choice : "has_choices"
    question ||--o{ vote : "receives_votes"
    
    vote ||--o{ vote_and_choice : "selects"
    choice ||--o{ vote_and_choice : "is_selected_in"

    user {
        UUID id PK
        VARCHAR name
        VARCHAR email UK
        TIMESTAMP created_at
    }

    user_password {
        UUID user_id PK,FK
        VARCHAR password_hash
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    vote_session {
        UUID id PK
        VARCHAR title
        TEXT description
        TIMESTAMP created_at
        TIMESTAMP ends_at
    }

    session_and_participant {
        UUID user_id PK,FK
        UUID session_id PK,FK
        TIMESTAMP invited_at
    }

    question {
        INT id PK
        UUID session_id FK
        VARCHAR text
        SMALLINT order_num
        BOOLEAN allow_multiple
        SMALLINT max_choices
        TIMESTAMP created_at
    }

    choice {
        INT id PK
        INT question_id FK
        VARCHAR text
        SMALLINT order_num
        TIMESTAMP created_at
    }

    vote {
        UUID id PK
        UUID user_id FK
        UUID session_id FK
        INT question_id FK
        TIMESTAMP created_at
    }

    vote_and_choice {
        UUID vote_id PK,FK
        INT choice_id PK,FK
    }

    user_history {
        UUID id PK
        UUID user_id FK
        UUID session_id FK
        SMALLINT version
        INT string_size
        BLOB receipt_data
        VARCHAR checksum
        TIMESTAMP created_at
    }

    result_history {
        UUID id PK
        UUID session_id FK
        SMALLINT version
        INT string_size
        BLOB result_data
        VARCHAR checksum
        TIMESTAMP created_at
    }
```

## Notes

### Migration SQLite → PostgreSQL

#### Changements nécessaires :
- TEXT (UUID) → UUID natif
- TEXT (timestamp) → TIMESTAMPTZ
- INTEGER (boolean) → BOOLEAN
- INTEGER AUTOINCREMENT → SERIAL
- Ajouter DEFAULT gen_random_uuid() pour les UUID
- Ajouter DEFAULT NOW() pour les timestamps