#!/bin/bash

# Quick development script for Aura CLI

set -e

case "$1" in
    "build")
        echo "Building Aura CLI..."
        make build
        ;;
    "test")
        echo "Running tests..."
        make test
        ;;
    "lint")
        echo "Running linter..."
        make lint
        ;;
    "format")
        echo "Formatting code..."
        make format
        ;;
    "dev")
        echo "Starting development server with hot reload..."
        air
        ;;
    "docker")
        echo "Starting Docker development environment..."
        docker-compose up -d aura-dev
        docker-compose exec aura-dev bash
        ;;
    *)
        echo "Usage: $0 {build|test|lint|format|dev|docker}"
        echo ""
        echo "Commands:"
        echo "  build   - Build the binary"
        echo "  test    - Run tests"
        echo "  lint    - Run linter"
        echo "  format  - Format code"
        echo "  dev     - Start development with hot reload"
        echo "  docker  - Start Docker development environment"
        exit 1
        ;;
esac
