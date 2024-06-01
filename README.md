# lissp

LIghtweight Stateful Sip Proxy

Proxy SIP traffic between 2 SIP UAs (no RTP). Implements RFC3261 specifications of a stateful proxy without maintining any state in memory, making it safe to reload and self-healing while maintining a minimal memory footprint.
