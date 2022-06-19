# GoSketch

Helper tool for quick sketch. 

Start db: docker compose-up -d

cd migrations &&  migrate --source=file://. --database=postgres://postgres:pass@localhost:5432/sketch\?sslmode=disable up  
