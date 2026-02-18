# Context for LLM – voting-app project (v1 MVP)

## Objectif
Monolithe modulaire Go + SQLite (shared DB), inspiré TDL v2.4 (Wild Workouts) sans chi/logrus.  
Découplage fort entre bounded contexts (users, sessions, questions).  
Pas de gRPC/RabbitMQ pour v1 – intraprocess + chan si async plus tard.  
Wiring hybride : décentralisé per service (dev) + central cmd/app/main.go (prod).

## Règles strictes à respecter
- Domain pur : entités + méthodes métier + Rehydrate (invariants). Pas d'infra (db.Timestamp, *sql.DB).
- Interfaces Repository dans app/ (locales, minimales).
- Conversions (toDto, To<Entity>, Rehydrate) uniquement dans adapters.
- Pas de cross-imports entre bounded contexts (même dans tests).
- Inbound ports : HTTP simple (net/http pur + custom router avec Group/Chain).
- Validation : structs Request + méthode Validate() dans ports.
- Logging : slog (init global ou NewLogger).
- Tests : isolés (uuid.New() pour IDs fictifs, memory repos si besoin).
- YAGNI : pas d'abstraction inutile (pas de common/db.Repository).

## État actuel (février 2026)
- Timestamp confiné adapters.
- Rehydrate partout.
- SessionChecker intraprocess pour découplage questions/sessions.
- Router custom (Group + Chain middleware).
- Schéma SQL centralisé (sql/schema.sql + go:embed).
- FK cross-BC supprimées (tests isolés).

## Ce que je veux éviter
- Couplage domain-infra.
- Imports cross-context.
- Mocks lourds (préférence tests directs ou memory repos).
- Chi/logrus/validator (sauf si vraiment justifié).

## Questions fréquentes à anticiper
- Pourquoi pas chi ? → Net/http pur pour zéro deps.
- Pourquoi pas gRPC ? → YAGNI v1.
- Pourquoi shared DB ? → Pragmatisme MVP (monolith-vs-micro article).
- Pourquoi Rehydrate ? → Reconstruction entité + invariants depuis DB.

Merci de rester fidèle à ces règles et de me dire quand je fais fausse route.