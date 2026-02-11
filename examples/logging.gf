-- logging.gf: Logging with std.log

import std.log

-- Default level is "info", so trace and debug are suppressed
log.trace("This trace message is hidden by default")
log.debug("This debug message is hidden by default")
log.info("Application started")
log.warn("Disk usage at 85%")
log.error("Connection to database failed")

-- Change log level to see more detail
log.setLevel("debug")
log.debug("Now debug messages are visible")
log.trace("But trace is still hidden")

-- Set to trace to see everything
log.setLevel("trace")
log.trace("Now trace is visible too")

-- Set back to warn to only see warnings and errors
log.setLevel("warn")
log.info("This info message is now hidden")
log.warn("Only warnings and errors show")
log.error("Errors always show at warn level")

println("Logging example complete")
