const HtmlWebpackPlugin = require("html-webpack-plugin");
const path = require("path");
const { globSync } = require("glob");
const protoPath = path.resolve(__dirname, '../protos/');

module.exports = {
    entry: "index.tsx",
    mode: "development",
    devtool: false,
    output: {
        filename: "main.js",
        path: path.resolve(__dirname, "dist"),
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: "index.html",
        }),
    ],
    resolve: {
        modules: [__dirname, protoPath, "App", "node_modules", path.resolve(__dirname, "./node_modules")],
        extensions: [".js", ".jsx", ".tsx", ".ts", ".proto", ".module.css", ".css", "..."],
    },
    module: {
        rules: [
            {
                test: /\.(js|ts)x?$/,
                exclude: /node_modules/,
                use: ["babel-loader"]
            },
            {
                test: /\.css$/,
                use: [
                    "style-loader",
                    "@teamsupercell/typings-for-css-modules-loader",
                    {
                        loader: "css-loader",
                        options: {
                            importLoaders: 1,
                            modules: {
                                localIdentName: '[path]__[name]__[local]__[hash:base64:5]'
                            }
                        }
                    },
                ]
            },
            {
                test: /\.(png|svg|jpg|gif)$/,
                exclude: /node_modules/,
                use: ["file-loader"]
            },
            {
                test: /\.proto$/,
                exclude: /node_modules/,
                use: {
                    loader: 'protobufjs-loader',
                    options: {
                        paths: [...globSync(path.resolve(protoPath, './*.proto')), protoPath],
                        pbjsArgs: [
                            '--null-semantics'
                        ],
                        pbts: {
                            output: protobufFile =>
                                path.join(__dirname, './App/protos/', path.basename(protobufFile) + '.d.ts')
                        },
                        target: 'static-module',
                    },
                },
            },
        ],
    },
};
