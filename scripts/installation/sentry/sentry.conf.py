# Sentry Configuration Template
# Environment variables will be substituted by envsubst

# This file is just Python, with a touch of Django which means
# you can inherit and tweak settings to your heart's content.

from sentry.conf.server import *  # NOQA

import os.path

CONF_ROOT = os.path.dirname(__file__)

# Database Configuration
DATABASES = {
    'default': {
        'ENGINE': 'sentry.db.postgres',
        'NAME': '${PROJ_SENTRY_POSTGRES_DB}',
        'USER': '${PROJ_SENTRY_POSTGRES_USER}',
        'PASSWORD': '${PROJ_SENTRY_POSTGRES_PASSWORD}',
        'HOST': '${PROJ_SENTRY_HOST}',
        'PORT': '5432',
        'AUTOCOMMIT': True,
        'ATOMIC_REQUESTS': False,
    }
}

# Redis Configuration for caching and queues
SENTRY_CACHE = 'sentry.cache.redis.RedisCache'
SENTRY_REDIS_OPTIONS = {
    'hosts': {
        0: {
            'host': '${PROJ_SENTRY_HOST}',
            'port': 6379,
            'db': 0,
        }
    }
}

# Queue Configuration
BROKER_URL = 'redis://${PROJ_SENTRY_HOST}:6379/0'

# Cache Configuration
CACHES = {
    'default': {
        'BACKEND': 'django_redis.cache.RedisCache',
        'LOCATION': 'redis://${PROJ_SENTRY_HOST}:6379/1',
        'OPTIONS': {
            'CLIENT_CLASS': 'django_redis.client.DefaultClient',
        }
    }
}

# Session Configuration
SESSION_ENGINE = 'django.contrib.sessions.backends.cache'
SESSION_CACHE_ALIAS = 'default'

# Security Configuration
SECRET_KEY = '${PROJ_SENTRY_SECRET_KEY}'

# Web Server Configuration
SENTRY_WEB_HOST = '${PROJ_SENTRY_HOST}'
SENTRY_WEB_PORT = ${PROJ_SENTRY_WEB_PORT}
SENTRY_WEB_OPTIONS = {
    'http': '%s:%s' % (SENTRY_WEB_HOST, SENTRY_WEB_PORT),
    'protocol': 'uwsgi',
    'uwsgi-socket': None,
    'http-keepalive': True,
    'memory-report': False,
    'processes': 1,
    'threads': 4,
}

# Allowed Hosts
ALLOWED_HOSTS = ['*']

# Email Configuration (optional)
EMAIL_BACKEND = 'django.core.mail.backends.smtp.EmailBackend'
EMAIL_HOST = 'localhost'
EMAIL_PORT = 587
EMAIL_USE_TLS = True
# EMAIL_HOST_USER = 'your-email@example.com'
# EMAIL_HOST_PASSWORD = 'your-email-password'
DEFAULT_FROM_EMAIL = 'sentry@${PROJ_SENTRY_HOST}'
SERVER_EMAIL = 'sentry@${PROJ_SENTRY_HOST}'

# File Storage Configuration
SENTRY_FILESTORE = 'django.core.files.storage.FileSystemStorage'
SENTRY_FILESTORE_OPTIONS = {
    'location': '/data/files',
}

# Logging Configuration
LOGGING = {
    'version': 1,
    'disable_existing_loggers': True,
    'handlers': {
        'console': {
            'level': 'INFO',
            'class': 'logging.StreamHandler',
            'formatter': 'simple',
        },
        'sentry': {
            'level': 'ERROR',
            'class': 'sentry.logging.handlers.SentryHandler',
        },
    },
    'formatters': {
        'simple': {
            'format': '%(levelname)s %(asctime)s %(module)s %(process)d %(thread)d %(message)s'
        }
    },
    'root': {
        'level': 'INFO',
        'handlers': ['console'],
    },
    'logger_overrides': {
        'sentry.errors': {
            'level': 'INFO',
            'handlers': ['console'],
            'propagate': False,
        },
    }
}

# Feature Flags
SENTRY_FEATURES = {
    'auth:register': True,
    'organizations:create': True,
    'projects:sample-events': True,
    'organizations:event-attachments': True,
    'organizations:performance-view': True,
    'organizations:discover-basic': True,
    'organizations:discover-query': True,
}

# Rate Limiting
SENTRY_RATELIMITER = 'sentry.ratelimits.redis.RedisRateLimiter'
SENTRY_RATELIMITER_OPTIONS = {
    'hosts': {
        0: {
            'host': '${PROJ_SENTRY_HOST}',
            'port': 6379,
            'db': 2,
        }
    }
}

# Node Storage
SENTRY_NODESTORE = 'sentry.nodestore.django.DjangoNodeStorage'

# Search Backend
SENTRY_SEARCH = 'sentry.search.django.DjangoSearchBackend'

# Time Zone
TIME_ZONE = 'UTC'

# Language Code
LANGUAGE_CODE = 'en-us'

# Debug Mode (disable in production)
DEBUG = False

# Template Debug
TEMPLATE_DEBUG = DEBUG

# Security Settings
SECURE_PROXY_SSL_HEADER = ('HTTP_X_FORWARDED_PROTO', 'https')
SECURE_SSL_REDIRECT = False
SESSION_COOKIE_SECURE = False
CSRF_COOKIE_SECURE = False

# CORS Settings
CORS_ORIGIN_ALLOW_ALL = False
CORS_ALLOW_CREDENTIALS = True

# Buffer Configuration
SENTRY_BUFFER = 'sentry.buffer.redis.RedisBuffer'
SENTRY_BUFFER_OPTIONS = {
    'hosts': {
        0: {
            'host': '${PROJ_SENTRY_HOST}',
            'port': 6379,
            'db': 3,
        }
    }
}

# Quotas Configuration
SENTRY_QUOTAS = 'sentry.quotas.redis.RedisQuota'
SENTRY_QUOTA_OPTIONS = {
    'hosts': {
        0: {
            'host': '${PROJ_SENTRY_HOST}',
            'port': 6379,
            'db': 4,
        }
    }
}

# Digests Configuration
SENTRY_DIGESTS = 'sentry.digests.backends.redis.RedisBackend'
SENTRY_DIGESTS_OPTIONS = {
    'hosts': {
        0: {
            'host': '${PROJ_SENTRY_HOST}',
            'port': 6379,
            'db': 5,
        }
    }
}

# Event Stream Configuration
SENTRY_EVENTSTREAM = 'sentry.eventstream.kafka.KafkaEventStream'
SENTRY_EVENTSTREAM_OPTIONS = {
    'cluster_options': {
        'bootstrap.servers': '${PROJ_SENTRY_HOST}:9092',
    },
    'producer_options': {
        'acks': 1,
    }
}

# Symbol Server Configuration
SENTRY_SYMBOLSERVER_OPTIONS = {
    'url': 'http://${PROJ_SENTRY_HOST}:3000',
}

# Data Scrubbing
SENTRY_SCRUB_DEFAULTS = True
SENTRY_SCRUB_IP_ADDRESS = True

# Sample Rate
SENTRY_SAMPLE_DATA = True

# Organization and Project Limits
SENTRY_SINGLE_ORGANIZATION = False
SENTRY_ALLOW_ORIGIN = '*'

# Authentication
AUTH_USER_MODEL = 'sentry.User'

# Social Authentication (optional)
# SOCIAL_AUTH_GOOGLE_OAUTH2_KEY = ''
# SOCIAL_AUTH_GOOGLE_OAUTH2_SECRET = ''
# SOCIAL_AUTH_GITHUB_KEY = ''
# SOCIAL_AUTH_GITHUB_SECRET = ''

# Integration Configuration
SENTRY_INTEGRATIONS = {
    'github': {
        'client_id': '',
        'client_secret': '',
    },
    'slack': {
        'client_id': '',
        'client_secret': '',
        'verification_token': '',
    },
    'jira': {
        'consumer_key': '',
        'private_key_path': '',
    }
}

# Performance Monitoring
SENTRY_PERFORMANCE_SAMPLE_RATE = 0.1
SENTRY_PERFORMANCE_MAX_SPANS = 1000

# Release Health
SENTRY_RELEASE_HEALTH_SAMPLE_RATE = 1.0