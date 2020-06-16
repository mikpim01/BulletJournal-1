import React, {useEffect} from 'react';

import './search.styles.less';
import {Empty, Result} from 'antd';
import {IState} from "../../store";
import {connect} from "react-redux";
import {search} from "../../features/search/action";
import {RouteComponentProps, withRouter} from "react-router";
import {useHistory, useParams} from "react-router-dom";
import {SearchResult} from "../../features/search/interface";
import {Loading} from "../../App";

type SearchProps = {
    searchPageNo: number;
    searchResult: SearchResult | undefined;
    searching: boolean;
    search: (term: string, scrollId?: string) => void;
};

const SearchPage: React.FC<SearchProps & RouteComponentProps> =
    ({
         searchPageNo,
         searchResult,
         searching,
         search
     }) => {
        const history = useHistory();
        // get id of note from router
        const {term} = useParams();

        useEffect(() => {
            if (!term) {
                return;
            }
            if (searchResult && searchResult.scrollId) {
                search(term, searchResult.scrollId);
            } else {
                search(term);
            }
        }, [term]);

        if (searching) {
            return <div className='search-page'><Loading/></div>;
        }

        if (!searchResult || searchResult.totalHits === 0) {
            return <div className='search-page'><Empty
                description={`No result found for "${term}"`}
            /></div>
        }

        return (
            <div className='search-page'>
                <div>
                    <Result
                        status='success'
                        title={`${searchResult.totalHits} results found for "${term}"`}
                    />
                </div>
            </div>
        );
    };

const mapStateToProps = (state: IState) => ({
    searchPageNo: state.search.searchPageNo,
    searchResult: state.search.searchResult,
    searching: state.search.searching
});

export default connect(mapStateToProps, {search})(withRouter(SearchPage));
