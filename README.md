# Einstein

PoC Go C2 server that mimics [Mattermost](https://mattermost.com/) HTTP traffic
## Important Things To Note

- Untested and unpolished code
- No help menu
- No exception handling
- No corresponding client provided
- No functionality to customize Einstein's configuration
- Filenames don't really mean anything
- Intentionally free of external packages/dependencies

## Running

Clone the repository and `go run Einstein.go`.

## Details

### Commands

| Command       | Purpose                          |
|---------------|----------------------------------|
| `list`        | List registered clients          |
| `select`      | Select client                    |
| `unselect`    | Unselect client                  |
| `current`     | Print currently-selected client  |
| `up`          | Upload file to client            |
| `ex`          | Execute system command on client |
| `persist`     | Install persistence on client    |
| `down`        | Download file from client        |
| `snap`        | Take screenshot of client        |
| `exit`/`quit` | Shutdown server                  |

### Endpoints

| Method | Path                                                        | Purpose                       |
|--------|-------------------------------------------------------------|-------------------------------|
| GET    | `/api/v<>/plugins/webapp`                                   | Pulse/Heartbeat               |
| POST   | `/api/v<>/users/status/ids`                                 | Retrieve job for client       |
| POST   | `/api/v<>/users/ids?since=<unix_timestamp_in_milliseconds>` | Retrieve response from client |
| GET    | `/static/<>.js`                                             | File transfer                 |

### Misc

- Client's unique ID and name is transmitted through a combination of the `MMAUTHTOKEN`, `MMUSERID`, and `MMCSRF` HTTP Cookie values
- Data transmitted is usually AES-128-GCM-encrypted and then ZBase32-encoded except in the case of file transfers where it is Base64-encoded after encryption instead
    - For file transfers, the final encoded data is smuggled via a minified JavaScript file used by Mattermost
    - For responses to jobs, the final encoded data is split into chunks across multiple Mattermost `status` structures where the data resides in the `UserId` field and the `Status` field will discreetly indicate the type of job to the client 