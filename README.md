# Einstein

PoC Go C2 server that mimics [Mattermost](https://mattermost.com/) HTTP traffic
## Important Things To Note

- Untested and unpolished code
- No exception handling
- No corresponding client provided
- Filenames don't really mean anything
- Intentionally free of external packages/dependencies

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
| `exit`/`quit` | Shutdown server                      |

### Endpoints

| Method | Path                                                        | Purpose                       |
|--------|-------------------------------------------------------------|-------------------------------|
| GET    | `/api/v<>/plugins/webapp`                                   | Pulse/Heartbeat               |
| POST   | `/api/v<>/users/status/ids`                                 | Retrieve job for client       |
| POST   | `/api/v<>/users/ids?since=<unix_timestamp_in_milliseconds>` | Retrieve response from client |
| GET    | `/static/<>.js`                                             | File transfer                 |