# Grafana Plugin For Disruptive backups

A <b>data-source backend plugin</b> that is used alongside [exprom-disruptive-backup](https://github.com/Hathoute/exprom-disruptive-backup) to display real-time metrics or old data. 

Using [Go](https://go.dev/) for backend and [ReactJS](https://fr.reactjs.org/) for frontend.

## Getting started

A data source backend plugin consists of both frontend and backend components.

### Frontend

1. Install dependencies

   ```bash
   yarn install
   ```

2. Build plugin in development mode or run in watch mode

   ```bash
   yarn dev
   ```

   or

   ```bash
   yarn watch
   ```

3. Build plugin in production mode

   ```bash
   yarn build
   ```

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v     # "go run mage.go -v" if you don't want to depend on mage
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

## License

This project is licensed under [Apache License 2.0](LICENSE) - see the [LICENSE](LICENSE) file for details

## Learn more

- [Build a data source backend plugin tutorial](https://grafana.com/tutorials/build-a-data-source-backend-plugin)
- [Grafana documentation](https://grafana.com/docs/)
- [Grafana Tutorials](https://grafana.com/tutorials/) - Grafana Tutorials are step-by-step guides that help you make the most of Grafana
- [Grafana UI Library](https://developers.grafana.com/ui) - UI components to help you build interfaces using Grafana Design System
- [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/)
