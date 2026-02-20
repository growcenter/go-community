# Event Session Registration Form Configuration - Explained

## Your Confusion Points

You're confused about:

1. **Lines 58-60**: `AdditionalRegistrantFormMode` and `AdditionalRegistrantFormID`
2. **Lines 71-72**: `RegistrationFormID`
3. **How to handle many custom questions**

Let me clarify the design from your spec document.

---

## Understanding the Two-Tier Form System

### The Problem Being Solved

When someone registers for an event session, there are **TWO types of registrants**:

1. **Primary Registrant** - The person doing the registration (themselves)
2. **Additional Registrants** - Other people they're registering (family, friends, guests)

**Example:** John registers for Christmas Service and brings his wife and 2 kids.

- **Primary**: John (has account, fills full form)
- **Additional**: Wife, Kid 1, Kid 2 (may not have accounts, simpler form)

---

## Field Breakdown

### 1. `RegistrationFormID` (Line 71-72)

```go
// Custom Questions (for primary registrant)
RegistrationFormID *int `gorm:"type:bigint" json:"registration_form_id"`
```

**Purpose:** References the **custom form** that the **PRIMARY registrant** must fill out.

**What it contains:**

- Custom questions specific to this session
- Examples: "Dietary preferences?", "T-shirt size?", "How did you hear about us?"
- This is a reference to the `forms` table (your universal form system)

**Example:**

```json
{
  "registration_form_id": 123 // References form with questions about dietary needs, t-shirt size, etc.
}
```

---

### 2. `AdditionalRegistrantFormMode` (Line 58-60)

```go
// Form Requirements for Additional Registrants
AdditionalRegistrantFormMode string `gorm:"type:varchar(30);default:'name_only'" json:"additional_registrant_form_mode"` // same_as_primary, name_only, custom
AdditionalRegistrantFormID   *int   `gorm:"type:bigint" json:"additional_registrant_form_id"`
```

**Purpose:** Defines **what information** is required for **ADDITIONAL registrants** (not the primary person).

**Three Options:**

#### Option 1: `name_only` (Default - Simplest)

- **Additional registrants only need to provide their NAME**
- No email, no phone, no custom questions
- **Use case:** Family Christmas service - just need names for headcount

**Example:**

```json
{
  "additional_registrant_form_mode": "name_only"
}
```

**Registration flow:**

```
Primary (John):
  - Name: John Doe ✓
  - Email: john@example.com ✓
  - Phone: +62812345678 ✓
  - Dietary preference: Vegetarian ✓
  - T-shirt size: L ✓

Additional (Wife):
  - Name: Jane Doe ✓
  (That's it! No email, no custom questions)

Additional (Kid 1):
  - Name: Jimmy Doe ✓
  (That's it!)
```

---

#### Option 2: `same_as_primary`

- **Additional registrants fill out THE SAME FORM as the primary registrant**
- Same custom questions, same identifier requirements
- **Use case:** Conference where everyone needs dietary preferences and t-shirt sizes

**Example:**

```json
{
  "additional_registrant_form_mode": "same_as_primary",
  "registration_form_id": 123 // Both primary and additional use this form
}
```

**Registration flow:**

```
Primary (John):
  - Name: John Doe ✓
  - Email: john@example.com ✓
  - Dietary preference: Vegetarian ✓
  - T-shirt size: L ✓

Additional (Wife):
  - Name: Jane Doe ✓
  - Email: jane@example.com ✓
  - Dietary preference: Vegan ✓
  - T-shirt size: M ✓
  (Same questions as primary!)
```

---

#### Option 3: `custom`

- **Additional registrants fill out a DIFFERENT FORM**
- Use `additional_registrant_form_id` to specify which form
- **Use case:** Kids program where kids need different questions than adults

**Example:**

```json
{
  "additional_registrant_form_mode": "custom",
  "registration_form_id": 123, // Adult form (for primary)
  "additional_registrant_form_id": 456 // Kids form (for additional)
}
```

**Registration flow:**

```
Primary (Parent):
  - Name: John Doe ✓
  - Email: john@example.com ✓
  - Dietary preference: Vegetarian ✓
  - Emergency contact: +62812345678 ✓

Additional (Kid):
  - Name: Jimmy Doe ✓
  - Parent email: john@example.com ✓
  - Age: 8 ✓
  - Allergies: Peanuts ✓
  (Different questions! Uses form 456)
```

---

## How to Handle Many Custom Questions?

### Answer: Use the Universal Form System

You **DON'T** add custom question fields directly to `event_session_model.go`. Instead:

1. **Create a form** in the `forms` table
2. **Add fields** to the `form_fields` table
3. **Reference the form** using `registration_form_id`

### Example: Creating a Form with Many Questions

```sql
-- Step 1: Create the form
INSERT INTO forms (code, title, form_type, created_by_community_id)
VALUES ('FORM-CHRISTMAS-2026', 'Christmas Service Registration', 'event_registration', 'admin123')
RETURNING id; -- Returns 789

-- Step 2: Add many custom questions
INSERT INTO form_fields (form_id, field_key, field_type, label, display_order, is_required) VALUES
  (789, 'dietary_preference', 'multiselect', 'Dietary Restrictions', 1, false),
  (789, 'tshirt_size', 'select', 'T-shirt Size', 2, true),
  (789, 'how_heard', 'text', 'How did you hear about us?', 3, false),
  (789, 'special_needs', 'textarea', 'Any special needs?', 4, false),
  (789, 'volunteer_interest', 'checkbox', 'Interested in volunteering?', 5, false),
  (789, 'transportation_needed', 'radio', 'Need transportation?', 6, false),
  (789, 'preferred_service_time', 'select', 'Preferred service time', 7, true),
  (789, 'bringing_guests', 'number', 'How many guests?', 8, false),
  (789, 'parking_needed', 'checkbox', 'Need parking?', 9, false),
  (789, 'accessibility_needs', 'textarea', 'Accessibility requirements?', 10, false);

-- Step 3: Assign to session
UPDATE event_sessions
SET registration_form_id = 789
WHERE code = 'SESSION-CHRISTMAS-2026';
```

Now when someone registers for this session, they'll see all 10 custom questions!

---

## Complete Real-World Example

### Scenario: Christmas Service with Family Registration

```go
// Session configuration
session := &EventSession{
    Code: "SESSION-CHRISTMAS-2026",
    Title: "Christmas Eve Service",

    // Registration rules
    RegistrationMode: "self_and_others",  // Can register family
    MaxRegistrationsPerUser: 10,           // Up to 10 people

    // PRIMARY registrant form (10 custom questions)
    RegistrationFormID: &formID789,  // Points to form with 10 questions

    // ADDITIONAL registrants (family members)
    AdditionalRegistrantFormMode: "name_only",  // Just need names
    AdditionalRegistrantFormID: nil,            // Not used when mode is "name_only"

    // Identifier requirements
    IdentifierConfig: JSONB{
        "primary": {
            "name": {"visible": true, "required": true},
            "email": {"visible": true, "required": true},
            "phone": {"visible": true, "required": true},
        },
        "additional": {
            "name": {"visible": true, "required": true},
            "email": {"visible": false, "required": false},  // Don't ask for email
            "phone": {"visible": false, "required": false},  // Don't ask for phone
        }
    },
}
```

**What happens when John registers:**

```
=== PRIMARY REGISTRANT (John) ===
Basic Info:
  ✓ Name: John Doe
  ✓ Email: john@example.com
  ✓ Phone: +62812345678

Custom Questions (Form 789):
  ✓ Dietary Restrictions: Vegetarian
  ✓ T-shirt Size: L
  ✓ How did you hear about us?: Facebook
  ✓ Special needs: None
  ✓ Interested in volunteering?: Yes
  ✓ Need transportation?: No
  ✓ Preferred service time: 6:00 PM
  ✓ How many guests?: 3
  ✓ Need parking?: Yes
  ✓ Accessibility requirements: None

=== ADDITIONAL REGISTRANT 1 (Wife) ===
Basic Info:
  ✓ Name: Jane Doe
  (No email, no phone, no custom questions!)

=== ADDITIONAL REGISTRANT 2 (Kid 1) ===
Basic Info:
  ✓ Name: Jimmy Doe
  (No email, no phone, no custom questions!)

=== ADDITIONAL REGISTRANT 3 (Kid 2) ===
Basic Info:
  ✓ Name: Jenny Doe
  (No email, no phone, no custom questions!)
```

---

## Summary Table

| Field                          | Purpose                                              | When to Use                           |
| ------------------------------ | ---------------------------------------------------- | ------------------------------------- |
| `RegistrationFormID`           | Custom questions for **PRIMARY** registrant          | Always (if you have custom questions) |
| `AdditionalRegistrantFormMode` | What info to collect from **ADDITIONAL** registrants | Always (defines the policy)           |
| `AdditionalRegistrantFormID`   | Custom form for **ADDITIONAL** registrants           | Only when `mode = "custom"`           |

---

## Decision Tree: Which Mode to Use?

```
Do additional registrants need custom questions?
├─ NO → Use "name_only"
│   └─ Example: Family Christmas service
│
└─ YES → Do they need the SAME questions as primary?
    ├─ YES → Use "same_as_primary"
    │   └─ Example: Conference (everyone needs dietary prefs)
    │
    └─ NO → Use "custom"
        └─ Example: Kids program (kids need different questions than adults)
```

---

## Common Mistakes to Avoid

### ❌ WRONG: Adding custom question fields to the model

```go
// DON'T DO THIS!
type EventSession struct {
    DietaryPreference string
    TShirtSize string
    HowHeardAboutUs string
    // ... adding more fields for each question
}
```

### ✅ CORRECT: Use the form system

```go
// DO THIS!
type EventSession struct {
    RegistrationFormID *int  // References a form with many fields
}
```

---

## Questions?

**Q: What if I want to add 50 custom questions?**  
A: Create a form with 50 fields in `form_fields` table. Reference it via `RegistrationFormID`.

**Q: What if different sessions need different questions?**  
A: Create different forms, assign different `RegistrationFormID` to each session.

**Q: Can I make some questions required for some sessions but not others?**  
A: Yes! Use `form_field_overrides` JSONB field to override requirements per session.

**Q: Do I need `AdditionalRegistrantFormID` if mode is "name_only"?**  
A: No, it's ignored. Only used when mode is "custom".

---

## Next Steps

1. ✅ Keep `RegistrationFormID` - it's for primary registrant custom questions
2. ✅ Keep `AdditionalRegistrantFormMode` - it defines the policy
3. ✅ Keep `AdditionalRegistrantFormID` - it's for custom additional forms
4. ✅ Use the `forms` and `form_fields` tables for all custom questions
5. ✅ Never add custom question fields directly to the session model
