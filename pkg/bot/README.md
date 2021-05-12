# telegram bot

## bot screens

```
                       (*)
                        │
                        ▼
╔════════════════════════════════════════════════╗
║ WelcomeScreen                                  ║
╟────────────────────────────────────────────────╢
║ Welcome!                                       ║
║ This is a WAQI bot                             ║
║                                                ║
║ Send a location to get its current air quality ║
╟────────────────────────────────────────────────╢
║ [Send location]                                ║
╚════════════════════════════════════════════════╝
              │                 ▲            ▲
              │ ─── @location   │            │
              │                 │ ─── /back  │
              ▼                 │            │
   ╔═════════════════════════════════════╗   │
   ║ LocationScreen                      ║   │
   ╟─────────────────────────────────────╢   │
   ║ %LocationName%                      ║   │
   ║ Current air quaility: Good          ║   │
   ║ CO_2: %CO2%                         ║   │
   ╟─────────────────────────────────────╢   │
   ║ [Subscribe to this location]        ║   │
   ╚═════════════════════════════════════╝   │
            │                                │
            │ ─── /subscribe                 │
            │                                │ ─── /unsubscribe
            ▼                                │
   ╔════════════════════════════════════════════════════════════╗
   ║ SubscribedScreen                                           ║
   ╟────────────────────────────────────────────────────────────╢
   ║ Subscribed to %LocationName%                               ║
   ║ You will receive notification if air quaility here changes ║
   ╟────────────────────────────────────────────────────────────╢
   ║ [Unsubscribe]                                              ║
   ╚════════════════════════════════════════════════════════════╝
```
