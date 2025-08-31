<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.proj.redis</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/redis-server</string>
        <string>${PROJ_REDIS_CONFIG_DIR}/redis.conf</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <true/>
    
    <key>UserName</key>
    <string>${USER}</string>
    
    <key>GroupName</key>
    <string>staff</string>
    
    <key>WorkingDirectory</key>
    <string>${PROJ_REDIS_DATA_DIR}</string>
    
    <key>StandardOutPath</key>
    <string>${PROJ_REDIS_LOG_DIR}/redis.log</string>
    
    <key>StandardErrorPath</key>
    <string>${PROJ_REDIS_LOG_DIR}/redis.error.log</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>${HOME}</string>
        <key>PROJ_SERVICE_NAME</key>
        <string>redis</string>
        <key>PROJ_SERVICE_VERSION</key>
        <string>${REDIS_VERSION}</string>
    </dict>
    
    <key>ProcessType</key>
    <string>Background</string>
    
    <key>ExitTimeOut</key>
    <integer>30</integer>
</dict>
</plist>