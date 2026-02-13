package entity

import "time"

// Example struct
type Policyholder struct {
    ID        int64     `db:"policyholder_id" json:"policyholder_id"`
    Name      string    `db:"name" json:"name"`
    Email     string    `db:"email" json:"email"`
    Country   string    `db:"country_code" json:"country_code"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}
