import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';
import {fetchImages} from "./images";
import {ConfigModal} from "./Modal";


class AppCard extends React.Component {
    buildMetadata;

    render() {
        return (
            <div className="col-sm-3 py-2">
                <div
                    className={`card h-100 border-right-4 ${this.color()} ${this.danger(this.buildpacks()) ? "border-danger" : ""}`}
                    style={{borderWidth: "medium"}}>
                    <div className="card-body d-flex flex-column">
                        <h4>{this.props.name}</h4>
                        <h6>
                            Team:{this.props.namespace}
                        </h6>
                        <small>
                            {this.buildpacks().map((item, i) =>
                                <div key={i}
                                     className={this.danger([item]) ? "text-danger font-weight-bold" : ""}>{item.key}:{item.version}<br/>
                                </div>
                            )}
                        </small>
                        {this.spinner()}
                    </div>
                </div>
            </div>);
    }

    color() {
        if (this.props.status === "True") {
            return "bg-success"
        } else if (this.props.status === "False") {
            return "bg-danger"
        } else if (this.props.latestImage !== "") {
            return "bg-success"
        }

        return "bg-secondary"
    }

    spinner() {
        if (this.props.status !== "Unknown") {
            return null;
        }

        return (
            <div className="mt-auto">
                <br/>
                <div className="spinner-border" role="status">
                    <span className="sr-only">Building...</span>
                </div>
                &nbsp;&nbsp;{this.percent()}
            </div>
        );
    }

    percent() {
        return ((this.props.remaining / 9) * 100).toFixed(0) + "%"
    }

    buildpacks() {
        if (this.props.buildMetadata == null) {
            return []
        }
        return this.props.buildMetadata
    }

    danger(items) {
        for (let i = 0; i < items.length; i++) {
            if (items[i].key === this.props.vulnerable.buildpack && items[i].version === this.props.vulnerable.version) {
                return true
            }

        }
        return false
    }
}

class AppList extends React.Component {
    render() {
        return (
            <div className="row">
                {this.props.apps.map((app, i) => <AppCard key={i} {...app} vulnerable={this.props.vulnerable}/>)}
            </div>
        );
    }
}


class App extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            images: [],
            vulnerable: {}
        }
    }

    componentDidMount() {
        setInterval(() => fetchImages()
            .then(res => this.setState({
                images: res,
                vulnerable: this.state.vulnerable
            })).catch(err => console.log(err)), 1000)
    }

    render() {
        return (
            <>
                <div className="container-fluid">
                    <div className={"row"}>
                        <ConfigModal
                            setVulnerable={(vulnerable) => this.setState({
                                images: this.state.images,
                                vulnerable: vulnerable
                            })}
                            vulnerable={this.state.vulnerable}/>
                    </div>
                    <AppList apps={this.state.images} vulnerable={this.state.vulnerable}/>
                </div>
            </>
        );
    }
}

export default App;
