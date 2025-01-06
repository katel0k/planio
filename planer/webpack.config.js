const HtmlWebpackPlugin = require("html-webpack-plugin");
const path = require("path");
const { globSync } = require("glob");
const protoPath = path.resolve(__dirname, '../protos/');

module.exports = {
    entry: "./src/index.tsx",
    mode: "development",
    output: {
        filename: "main.js",
        path: path.resolve(__dirname, "dist"),
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: "./src/index.html",
        }),
    ],
    resolve: {
        modules: [__dirname, "src", "node_modules", protoPath, path.join(__dirname, 'node_modules')],
        extensions: [".*", ".js", ".jsx", ".tsx", ".ts", ".proto"],
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
                exclude: /node_modules/,
                use: ["style-loader", "css-loader"]
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
                        paths: globSync(path.resolve(protoPath, './*.proto')),
                        pbjsArgs: [],
                        pbts: {
                            outDir: path.join(__dirname, './src/protos/')
                        },
                        target: 'static-module',
                    },
                },
            },
        ],
    },
};
