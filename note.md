# Structure Inspiration

This structure is inspired by Django's architecture, where each module stands independently within the app directory. The same concept is applied here in Go, with several personal preferences. Key implementations include:

    Schema-based Testing: Each test is isolated within its own PostgreSQL schema.

    Modular Separation by App: Each domain or module is separated per app, similar to Django's app.

# Potential Improvements

    Logging Integration: Logs should be sent to tools like Grafana or Loki, rather than just being saved as JSON files.

    Audit Logging: Introduce an audit log library that can be integrated with the main logging system.

There are actually many more areas that can still be improved,
but even for a simple CRUD application, this setup is already somewhat over-engineered.