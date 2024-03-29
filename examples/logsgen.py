#!/usr/bin/env python3

from faker.cli import execute_from_command_line
from faker import Faker
import pystache
import argparse
import sys

try:
    xrange
except NameError:
    xrange = range

fake = Faker()

class FakerThing(object):
    def __getattr__(self, name):
        return fake.format(name)

if __name__ == '__main__':
	# execute_from_command_line()

	parser = argparse.ArgumentParser(description='Generate log files per template.')
	parser.add_argument('template', metavar='template', type=argparse.FileType('rt'), help='template file, use - for <stdin>')
	parser.add_argument('count', metavar='count', type=int, help='number of records to generate')
	parser.add_argument('output', metavar='output', type=argparse.FileType('wt'), nargs='?', help='result output file, use - for <stdout> (default)')
	parser.set_defaults(output='-')

	args = parser.parse_args()

	if args.count < 1:
		sys.stderr.write("Error: count parameter should be a positive number.")
		sys.exit(1)

	template = args.template.read()
	
	try:
		template = template.decode('utf-8')
	except AttributeError:
		pass

	parsed = pystache.parse(template)
	for x in xrange(0, args.count):
		args.output.write(pystache.render(parsed, FakerThing()))
