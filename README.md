# cats

```text
        _                        
       \`*-.                    
        )  _`-.                 
       .  : `. .                
       : _   '  \               
       ; *` _.   `*-._          
       `-.-'          `-.       
         ;       `       `.     
         :.       .        \    
         . \  .   :   .-'   .   
         '  `+.;  ;  '      :   
         :  '  |    ;       ;-. 
         ; '   : :`-:     _.`* ;
[bug] .*' /  .*' ; .*`- +'  `*' 
      `*-*   `*-*  `*-*'
```

A lightweight Go web server for serving random cat pictures.

## Endpoints

| Route | Method | Description |
| :--- | :--- | :--- |
| `/` | `GET` | Redirects (`302`) to a random cat image URL (`/cats/<filename>`) |
| `/cat` | `GET` | Directly streams a random cat image |
| `/cats/<filename>` | `GET` | Serves a specific cat image file from `images/` |
| `/api/cat` | `GET` | Returns JSON metadata for a random cat image |
| `/api/cats` | `GET` | Returns a JSON list of all available cat images |
| `/favicon.ico` | `GET` | Serves [`favicon.svg`](./favicon.svg) |
| `/healthz` | `GET` | Health check endpoint |

## Quick Start

### Local

```bash
go run .
```

The server listens on `http://localhost:8090` by default.

### Docker

```bash
docker build -t cats .
docker run -p 8090:8090 cats
```

## Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8090` | Port to bind the HTTP server to |
| `IMAGES_DIR` | `images` | Path to the directory containing cat images |

## Adding Images

Drop `.png`, `.jpg`, `.jpeg`, `.webp`, or `.gif` files into the [`images/`](./images/) folder and restart the server.
