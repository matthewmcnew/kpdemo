import React from 'react';
import './App.css';
import 'bootstrap/dist/css/bootstrap.css';
import {fetchImages} from "./images";
import {ConfigModal} from "./Modal";


class AppCard extends React.Component {
    buildMetadata;

    render() {
        return (
            <div className="col-sm-2 py-2">
                <div
                    className={`card h-100 border-right-4 ${this.color()} ${this.danger(this.buildpacks(), this.props.runImage) ? "border-danger" : ""}`}
                    style={{borderWidth: "medium"}}>
                    <div className="card-body d-flex flex-column">
                        <h5>{this.props.name}</h5>
                        <div className={"team-name"}>
                            Team:{this.props.namespace}
                        </div>
                        {this.runImage()}

                        {this.buildpacks().map((item, i) =>
                            <div key={i}
                                 className={this.danger([item]) ? "text-danger font-weight-bold vulnerable" : "not-vulnerable"}>{item.key}:{item.version}<br/>
                            </div>
                        )}

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
        if (this.props.remaining < 0) {
            return "Queued"
        }

        return ((this.props.completed / this.props.remaining) * 100).toFixed(0) + "%"
    }

    buildpacks() {
        if (this.props.buildMetadata == null) {
            return []
        }
        return this.props.buildMetadata
    }

    danger(items, runImage) {
        if (runImage !== undefined && this.props.vulnerable.runImage !== "" && runImage.includes(this.props.vulnerable.runImage)) {
            return true;
        }

        for (let i = 0; i < items.length; i++) {
            if (items[i].key === this.props.vulnerable.buildpack && items[i].version === this.props.vulnerable.version) {
                return true
            }

        }
        return false
    }

    runImage() {
        if (this.props.runImage === "") {
            return null
        }

        const runImageParts = this.props.runImage.replace('index.docker.io/', '').split("@")

        if (2 !== runImageParts.length) {
            return null
        }

        const vulnerable = this.props.runImage.includes(this.props.vulnerable.runImage)

        return (
            <>
                <div className={vulnerable ? "text-danger font-weight-bold vulnerable" : "not-vulnerable"}>
                    Stack:{runImageParts[0].substring(0, 20)} {runImageParts[1].substr(7, 6)}<br/>
                </div>
                <div className={"buildpack-divider"}/>
            </>
        );
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
