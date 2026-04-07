# Room Booking — Frontend

React + Vite frontend for the Room Booking API.

## Setup

```bash
npm install
npm run dev
```

App runs at **http://localhost:3000**. Backend must be at **http://localhost:8080** with CORS enabled.

## Project structure

```
src/
├── api/
│   └── client.js              # All API calls in one place
├── components/
│   ├── ui/
│   │   └── index.jsx          # Shared UI primitives (Loader, Alert, etc.)
│   ├── admin/
│   │   ├── RoomsTab.jsx       # Room list + create room + set schedule
│   │   └── BookingsTab.jsx    # Paginated table of all bookings
│   ├── user/
│   │   ├── BookRoomTab.jsx    # 4-step booking flow
│   │   └── MyBookingsTab.jsx  # Upcoming bookings + cancel
│   ├── modals/
│   │   ├── Modal.jsx          # Generic modal wrapper
│   │   ├── CreateRoomModal.jsx
│   │   └── CreateScheduleModal.jsx
│   └── Topbar.jsx             # Sticky header with tabs + role badge
├── context/
│   └── AuthContext.jsx        # Token + role state (sessionStorage)
├── pages/
│   └── LoginPage.jsx          # Role picker → dummyLogin
├── utils/
│   └── format.js              # Date/time formatters
├── App.jsx                    # Root — routing between admin/user views
├── App.module.css
├── index.css                  # Global CSS variables + base styles
└── main.jsx
```

## API endpoints used

| Method | Path | Who |
|--------|------|-----|
| POST | `/dummyLogin` | Login page |
| GET | `/rooms/list` | Admin & User |
| POST | `/rooms/create` | Admin |
| POST | `/rooms/:id/schedule/create` | Admin |
| GET | `/rooms/:id/slots/list?date=` | User |
| POST | `/bookings/create` | User |
| GET | `/bookings/my` | User |
| GET | `/bookings/list` | Admin |
| POST | `/bookings/:id/cancel` | User |

## CORS

Your Go/Node backend needs to allow requests from `http://localhost:3000`. Example for Go:

```go
w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
```
