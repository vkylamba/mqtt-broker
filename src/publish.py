#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright (c) 2014 Roger Light <roger@atchoo.org>
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Distribution License v1.0
# which accompanies this distribution.
#
# The Eclipse Distribution License is available at
#   http://www.eclipse.org/org/documents/edl-v10.php.
#
# Contributors:
#    Roger Light - initial implementation

# This shows an example of using the publish.multiple helper function.

import paho.mqtt.publish as publish

msgs = [
    {
        'topic': "/iotaapsys/services/heartbeat", 'payload': "multiple 1"
    },
    ("/iotaapsys/services/heartbeat", "multiple 2", 0, False)
]
publish.multiple(msgs, hostname="localhost", port=9024)
