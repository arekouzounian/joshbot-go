#!/usr/bin/python3

import sys
import logging
logging.basicConfig(stream=sys.stderr)

from api import create_app

application = create_app()